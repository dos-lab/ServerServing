package service

import (
	daModels "ServerServing/da/mysql/da_models"
	SErr "ServerServing/err"
	"ServerServing/internal/dal"
	"ServerServing/internal/internal_models"
	"fmt"
	"log"
)

type ServersService struct{}

type LoadServerAccountArg struct {
	// From Size 如果指定了这两个数值，都不为0，则使用它们。否则全部load全部的Account
	From uint
	Size uint
}

type LoadServerDetailArg struct {
	// WithHardwareInfo 指定是否加载硬件的元信息
	WithHardwareInfo bool
	// WithAccountArg 加载账户信息的参数，为nil则不加载
	WithAccountArg *LoadServerAccountArg
	// WithRemoteAccessUsages 指定是否加载正在远程登录这台服务器的用户信息。
	WithRemoteAccessUsages bool
	// WithGPUUsages 指定是否加载GPU的使用信息。
	WithGPUUsages bool
	// WithCPUMemProcessesUsageInfo 指定是否加载CPU，内存，进程的使用信息。
	WithCPUMemProcessesUsageInfo bool
}

func GetServersService() *ServersService {
	return &ServersService{}
}

// Info 获取一个Server数据。
func (s *ServersService) Info(Host string, Port uint, arg *LoadServerDetailArg) (*internal_models.ServerInfo, *SErr.APIErr){
	// 每个请求内部共享一次SSH session
	serverInfo := &internal_models.ServerInfo{}
	// 需要加载Basic的信息，与Account的加载同时进行。

	// 第一步，从MySQL加载初步的Server信息，获取该服务器的管理员账号与密码，以便后续的信息获取。
	serverBasic, accounts, err := s.loadBasic(Host, Port, arg.WithAccountArg)
	if err != nil {
		return nil, err
	}
	serverInfo.Basic = serverBasic
	// 从服务器中能够获取到一份Account数据，但是并不一定是最新的服务器中的用户数据。
	serverInfo.AccountInfos = &internal_models.ServerAccountInfos{
		Accounts:         accounts,
	}

	// 第二步，初始化到该服务器的连接，如果连接失败，则直接返回错误。
	es, err := s.getExecutorService(Host, Port, serverBasic.OSType, serverBasic.AdminAccountName, serverBasic.AdminAccountPwd)
	if err != nil {
		causeDescription := fmt.Sprintf("在尝试使用管理员账户与该服务器建连时失败！请检查该账户的配置以及网络状况！内嵌的出错信息为：[%s]", err.Message)
		log.Printf("ServersService Info，causeDescription=[%s]", causeDescription)
		serverInfo.AccessFailedInfo.CauseDescription = causeDescription
		return nil, SErr.SSHConnectionErr.CustomMessage(causeDescription)
	}

	// 接下来，分别对LoadServerDetailArg中的每个可选项进行针对性的load

	// 对WithAccountArg做load：
	// 账户信息在MySQL中存储一份，但是不一定准确（因为Server可能随时被人修改）
	// 所以，每当从MySQL查询出一份数据后，我们需要对它修正。具体修正逻辑为：
	// 我们将会从Server中查询一份实时的最新的用户信息（查不到密码）
	// 如果MySQL中，没有存储该账户的信息，则使用从Server查询的最新数据插入该用户的数据。
	// 如果MySQL存储了，并且从Server能够查询到该用户（一致的），则将它的信息进行补全。
	// 如果MYSQL存储了，但是从Server中查不到该用户（可能被删掉了），那么就把他的数据过滤掉（不在MySQL中删除）
	s.combineAccounts(es, arg, serverInfo)

	// 对WithHardwareInfo做load
	s.loadHardwareInfo(es, arg, serverInfo)

	// 对WithCPUMemProcessesUsageInfo做load
	s.loadCPUMemProcessesUsageInfo(es, arg, serverInfo)

	// 对WithGPUUsages做load
	s.loadGPUUsages(es, arg, serverInfo)

	return serverInfo, nil
}

func (s *ServersService) getExecutorService(Host string, Port uint, osType daModels.OSType, adminAccountName, adminAccountPwd string) (ExecutorService, *SErr.APIErr) {
	switch osType {
	case daModels.OSTypeLinux:
		return GetLinuxSSHExecutorService(Host, Port, adminAccountName, adminAccountPwd)
	default:
		panic("Unimplemented")
	}
}

// Infos 获取一批Server数据。目前所有Server使用同一个arg参数指定它对应的Detail信息量。
func (s *ServersService) Infos(from, size int, arg *LoadServerDetailArg, keyword *string) {

}

func (s *ServersService) combineAccounts(es ExecutorService, arg *LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	if arg.WithAccountArg == nil {
		return
	}
	serverInfo.AccountInfos = &internal_models.ServerAccountInfos{
		ServerInfoCommon: &internal_models.ServerInfoCommon{},
		Accounts:         make([]*internal_models.Account, 0),
	}
	getAccountListResp, err := es.GetAccountList()
	serverInfo.AccountInfos.Output = getAccountListResp.Output
	if err != nil {
		serverInfo.AccountInfos.FailedInfo = &internal_models.ServerInfoLoadingFailedInfo{
			CauseDescription: fmt.Sprintf("向服务器查询用户列表时出错，es=[%s], 出错信息为：[%s]", es, err.Error()),
		}
		return
	}
	accountsInServerMap := make(map[string]*internal_models.Account)
	for _, accountInServer := range getAccountListResp.Accounts {
		accountsInServerMap[accountInServer.Name] = accountInServer
	}

	for _, acc := range serverInfo.AccountInfos.Accounts {
		if accInServer, ok := accountsInServerMap[acc.Name]; ok {
			// 当从服务器中查询到了Account与从MySQL中一样的账户，则补全信息。
			// 在MySQL中没有存储的信息，将它们补全。
			acc.UID = accInServer.UID
			acc.GID = accInServer.GID
		} else {
			// 从MySQL中存储的账户不存在与在从服务器中存储的账户列表
			// 那么，即该账户可能是在Server中被擅自删除了，那么这时先给该账户打一个标记。
			acc.NotExistsInServer = true
		}
		delete(accountsInServerMap, acc.Name)
	}

	if len(accountsInServerMap) > 0 {
		// 剩余的在accountsInServerMap中，没有被遍历到的账户，
		// 则代表该账户在Server中存在，但是在MySQL中没有存储。
		// 所以将它们插入到MySQL中。
		toBeInserted := make([]*daModels.Account, 0, len(accountsInServerMap))
		for _, accountInServer := range accountsInServerMap {
			toBeInserted = append(toBeInserted, &daModels.Account{
				Name:      accountInServer.Name,
				Pwd:       accountInServer.Pwd,
				Host:      accountInServer.Host,
				Port:      accountInServer.Port,
			})
		}
		accDal := dal.GetAccountDal()
		err := accDal.Upsert(toBeInserted)
		if err != nil {
			// 非关键错误，仅仅打印log即可。
			log.Printf("ServersService Upsert 不在MySQL中的账户时失败！es=[%s], 错误为：[%s]", es, err.Error())
		}
		// 插入到MySQL后，将他们补全到serverInfo中。
		for _, accInServer := range accountsInServerMap {
			serverInfo.AccountInfos.Accounts = append(serverInfo.AccountInfos.Accounts, accInServer)
		}
	}
}

// loadHardwareInfo 加载硬件相关信息
func (s *ServersService) loadHardwareInfo(es ExecutorService, arg *LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	// TODO
}

// loadRemoteAccessUsages 加载正在远程访问该Server的用户使用信息。
func (s *ServersService) loadRemoteAccessUsages(es ExecutorService, arg *LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	// TODO
}

// loadGPUUsages 加载GPU的使用情况
func (s *ServersService) loadGPUUsages(es ExecutorService, arg *LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
    // TODO
}

// loadCPUMemProcessesUsageInfo 加载当前正在使用CPU，内存，以及进程的占用信息。
func (s *ServersService) loadCPUMemProcessesUsageInfo(es ExecutorService, arg *LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	if !arg.WithCPUMemProcessesUsageInfo {
		return
	}
	serverInfo.CPUMemProcessesUsageInfo = &internal_models.ServerCPUMemProcessesUsageInfo{
		ServerInfoCommon: &internal_models.ServerInfoCommon{},
	}
	topResp, err := es.Top()
	serverInfo.CPUMemProcessesUsageInfo.Output = topResp.Output
	if err != nil {
		serverInfo.CPUMemProcessesUsageInfo.FailedInfo = &internal_models.ServerInfoLoadingFailedInfo{
			CauseDescription: fmt.Sprintf("加载Top信息时失败！es=[%s]，出错信息为：[%s]", es, err),
		}
		return
	}
	serverInfo.CPUMemProcessesUsageInfo.CPUMemUsage = topResp.CPUMemUsage
	serverInfo.CPUMemProcessesUsageInfo.ProcessInfos = topResp.ProcessInfos
}

// loadBasic 从MySQL加载一个基本的Server信息，是所有其他的查询的基本步骤。
func (s *ServersService) loadBasic(Host string, Port uint, withAccountArg *LoadServerAccountArg) (*internal_models.ServerBasic, []*internal_models.Account, *SErr.APIErr) {
	serverDal := dal.GetServerDal()
	if withAccountArg.From == 0 && withAccountArg.Size == 0 {
		// 都为空时，全部加载
		server, err := serverDal.Get(Host, Port, true)
		if err != nil {
			return nil, nil, err
		}
		return s.packServer(server), s.packAccounts(server.Accounts), nil
	}
	server, err := serverDal.Get(Host, Port, false)
	if err != nil {
		return nil, nil, err
	}
	resServer := s.packServer(server)
	accountDal := dal.GetAccountDal()
	accs, _, err := accountDal.List(Host, Port, withAccountArg.From, withAccountArg.Size)
	if err != nil {
		return nil, s.packPtrAccounts(accs), err
	}
	return resServer, s.packPtrAccounts(accs), nil
}

func (s *ServersService) packServer(server *daModels.Server) *internal_models.ServerBasic {
	return &internal_models.ServerBasic{
		CreatedAt:  server.CreatedAt,
		UpdatedAt: server.UpdatedAt,
		DeletedAt: server.DeletedAt,
		Host:             server.Host,
		Port:             server.Port,
		AdminAccountName: server.AdminAccountName,
		AdminAccountPwd:  server.AdminAccountPwd,
	}
}

func (s *ServersService) packAccount(account *daModels.Account) *internal_models.Account {
	return &internal_models.Account{
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
		DeletedAt: account.DeletedAt,
		Name:      account.Name,
		Pwd:       account.Pwd,
		Host:      account.Host,
		Port:      account.Port,
	}
}

func (s *ServersService) packPtrAccounts(accounts []*daModels.Account) []*internal_models.Account {
	res := make([]*internal_models.Account, 0, len(accounts))
	for _, ac := range accounts {
		res = append(res, s.packAccount(ac))
	}
	return res
}

func (s *ServersService) packAccounts(accounts []daModels.Account) []*internal_models.Account {
	res := make([]*internal_models.Account, 0, len(accounts))
	for _, ac := range accounts {
		res = append(res, s.packAccount(&ac))
	}
	return res
}
