package service

import (
	daModels "ServerServing/da/mysql/da_models"
	SErr "ServerServing/err"
	"ServerServing/internal/dal"
	"ServerServing/internal/internal_models"
	"ServerServing/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"sync"
)

type ServersService struct {
}

func GetServersService() *ServersService {
	return &ServersService{}
}

func (s *ServersService) Create(c *gin.Context, Host string, Port uint, OSType daModels.OSType, adminAccountName, adminAccountPwd string) *SErr.APIErr {
	es, err := OpenExecutorService(Host, Port, OSType, adminAccountName, adminAccountPwd)
	if err != nil {
		return err
	}
	defer es.Close()
	// 能够联通该服务器，则调用MySQL创建。
	serverDal := dal.GetServerDal()
	err = serverDal.Create(&daModels.Server{
		Host:             Host,
		Port:             Port,
		AdminAccountName: adminAccountName,
		AdminAccountPwd:  adminAccountPwd,
		OSType:           OSType,
	})
	if err != nil {
		log.Printf("ServersService serverDal.Create failed, err=[%v]", err)
		return err
	}
	return nil
}

func (s *ServersService) Delete(c *gin.Context, Host string, Port uint) *SErr.APIErr {
	// 能够联通该服务器，则调用MySQL创建。
	serverDal := dal.GetServerDal()
	err := serverDal.Delete(Host, Port)
	if err != nil {
		log.Printf("ServersService serverDal.Create failed, err=[%v]", err)
		return err
	}
	return nil
}

// basicInfo 从MySQL获取服务器的基本数据
func (s *ServersService) basicInfo(c *gin.Context, Host string, Port uint) (*internal_models.ServerBasic, []*internal_models.ServerAccount, *SErr.APIErr) {
	// 每个请求内部共享一次SSH session
	// 需要加载Basic的信息，与Account的加载同时进行。
	// 从MySQL加载初步的Server信息，获取该服务器的管理员账号与密码，以便后续的信息获取。
	serverDal := dal.GetServerDal()
	daServer, err := serverDal.Get(Host, Port, true)
	if err != nil {
		return nil, nil, err
	}
	serverBasic, accounts := s.packServer(daServer), s.packAccounts(daServer.Accounts)
	return serverBasic, accounts, nil
}

// Info 获取一个Server数据。
func (s *ServersService) Info(c *gin.Context, Host string, Port uint, arg *internal_models.LoadServerDetailArg) (*internal_models.ServerInfo, *SErr.APIErr) {
	serverBasic, accounts, err := s.basicInfo(c, Host, Port)
	if err != nil {
		return nil, err
	}
	serverInfo := &internal_models.ServerInfo{}
	serverInfo.Basic = serverBasic
	// 从服务器中能够获取到一份Account数据，但是并不一定是最新的服务器中的用户数据。
	serverInfo.AccountInfos = &internal_models.ServerAccountInfos{
		Accounts: accounts,
	}
	// 第二步，初始化到该服务器的连接，如果连接失败，则直接返回错误。
	es, err := s.openExecutorService(serverBasic.Host, serverBasic.Port, serverBasic.OSType, serverBasic.AdminAccountName, serverBasic.AdminAccountPwd)
	defer func() {
		_ = es.Close()
	}()
	if err != nil {
		serverInfo.AccessFailedInfo.CauseDescription = err.Message
		return serverInfo, err
	}
	s.loadInfoFromServer(serverInfo, es, arg)
	return serverInfo, nil
}

func (s *ServersService) openExecutorService(Host string, Port uint, osType daModels.OSType, AdminAccountName, AdminAccountPwd string) (ExecutorService, *SErr.APIErr) {
	es, err := OpenExecutorService(Host, Port, osType, AdminAccountName, AdminAccountPwd)
	if err != nil {
		causeDescription := fmt.Sprintf("在尝试使用管理员账户与该服务器建连时失败！请检查该账户的配置以及网络状况！内嵌的出错信息为：[%s]", err.Message)
		log.Printf("ServersService Info，causeDescription=[%s]", causeDescription)
		return nil, SErr.SSHConnectionErr.CustomMessage(causeDescription)
	}
	return es, nil
}

func (s *ServersService) openExecutorServiceByHostPort(c *gin.Context, Host string, Port uint) (ExecutorService, *SErr.APIErr) {
	serverBasic, _, err := s.basicInfo(c, Host, Port)
	if err != nil {
		return nil, err
	}
	return s.openExecutorService(serverBasic.Host, serverBasic.Port, serverBasic.OSType, serverBasic.AdminAccountName, serverBasic.AdminAccountPwd)
}

func (s *ServersService) loadInfoFromServer(targetServerInfo *internal_models.ServerInfo, es ExecutorService, arg *internal_models.LoadServerDetailArg) {
	// 接下来，分别对LoadServerDetailArg中的每个可选项进行针对性的load
	// 对WithAccountArg做load：
	// 账户信息在MySQL中存储一份，但是不一定准确（因为Server可能随时被人修改）
	// 所以，每当从MySQL查询出一份数据后，我们需要对它修正。具体修正逻辑为：
	// 我们将会从Server中查询一份实时的最新的用户信息（查不到密码）
	// 如果MySQL中，没有存储该账户的信息，则使用从Server查询的最新数据插入该用户的数据。
	// 如果MySQL存储了，并且从Server能够查询到该用户（一致的），则将它的信息进行补全。
	// 如果MYSQL存储了，但是从Server中查不到该用户（可能被删掉了），那么就把他的数据过滤掉（不在MySQL中删除）
	s.loadAccounts(es, arg, targetServerInfo)
	// 对WithHardwareInfo做load：
	// 目前包含CPU和GPU的硬件数据。
	s.loadHardwareInfo(es, arg, targetServerInfo)
	// 对WithRemoteAccessUsages做load
	// 包含了当前正在使用远程访问该服务器的用户信息
	s.loadRemoteAccessUsages(es, arg, targetServerInfo)
	// 对WithCPUMemProcessesUsageInfo做load
	s.loadCPUMemProcessesUsageInfo(es, arg, targetServerInfo)
	// 对WithGPUUsages做load
	s.loadGPUUsages(es, arg, targetServerInfo)
}

// Infos 获取一批Server数据。目前所有Server使用同一个arg参数指定它对应的Detail信息量。
func (s *ServersService) Infos(c *gin.Context, from, size uint, arg *internal_models.LoadServerDetailArg, keyword *string) ([]*internal_models.ServerInfo, uint, *SErr.APIErr) {
	serverDal := dal.GetServerDal()
	var servers []*daModels.Server
	var total uint
	var err *SErr.APIErr
	if keyword != nil {
		servers, total, err = serverDal.SearchByHostAndAdmin(from, size, *keyword, arg.WithAccounts)
	} else {
		servers, total, err = serverDal.List(from, size, arg.WithAccounts)
	}
	if err != nil {
		return nil, 0, err
	}
	resultServerInfos := make([]*internal_models.ServerInfo, 0, len(servers))
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for _, daServer := range servers {
		daServer := daServer
		serverBasic, accounts := s.packServer(daServer), s.packAccounts(daServer.Accounts)
		util.GoWithWG(wg, func() {
			serverInfo := &internal_models.ServerInfo{}
			serverInfo.Basic = serverBasic
			// 从服务器中能够获取到一份Account数据，但是并不一定是最新的服务器中的用户数据。
			serverInfo.AccountInfos = &internal_models.ServerAccountInfos{
				Accounts: accounts,
			}
			// 第二步，初始化到该服务器的连接，如果连接失败，则直接返回错误。
			es, err := s.openExecutorService(serverBasic.Host, serverBasic.Port, serverBasic.OSType, serverBasic.AdminAccountName, serverBasic.AdminAccountPwd)
			defer func() {
				_ = es.Close()
			}()
			if err != nil {
				serverInfo.AccessFailedInfo.CauseDescription = err.Message
			} else {
				s.loadInfoFromServer(serverInfo, es, arg)
			}
			mu.Lock()
			defer mu.Unlock()
			resultServerInfos = append(resultServerInfos, serverInfo)
		})
	}
	wg.Wait()
	return resultServerInfos, total, nil
}

func (s *ServersService) loadAccounts(es ExecutorService, arg *internal_models.LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	if arg.WithAccounts == false {
		return
	}
	if serverInfo.AccountInfos == nil {
		serverInfo.AccountInfos = &internal_models.ServerAccountInfos{}
	}
	serverInfo.AccountInfos.ServerInfoCommon = &internal_models.ServerInfoCommon{}
	getAccountListResp, err := es.GetAccountList()
	serverInfo.AccountInfos.Output = getAccountListResp.Output
	if err != nil {
		serverInfo.AccountInfos.FailedInfo = &internal_models.ServerInfoLoadingFailedInfo{
			CauseDescription: fmt.Sprintf("向服务器查询用户列表时出错，es=[%s], 出错信息为：[%s]", es, err.Error()),
		}
		return
	}
	accountsInServerMap := make(map[string]*internal_models.ServerAccount)
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
			// 从MySQL中存储的账户不存在于在从服务器中存储的账户列表
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
				Name: accountInServer.Name,
				Pwd:  accountInServer.Pwd,
				Host: accountInServer.Host,
				Port: accountInServer.Port,
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

	s.loadAccountBackupDirInfos(es, arg, serverInfo)

}

func (s *ServersService) loadAccountBackupDirInfos(es ExecutorService, arg *internal_models.LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	if !arg.WithBackupDirInfo {
		return
	}
	wg := &sync.WaitGroup{}
	batch := 5
	for i, account := range serverInfo.AccountInfos.Accounts {
		account := account
		util.GoWithWG(wg, func() {
			account.BackupDirInfo = &internal_models.ServerAccountBackupDirInfo{
				ServerInfoCommon: &internal_models.ServerInfoCommon{
					Output:     "",
					FailedInfo: nil,
				},
			}
			resp, err := es.GetBackupDir(account.Name)
			if err != nil {
				account.BackupDirInfo.FailedInfo = &internal_models.ServerInfoLoadingFailedInfo{CauseDescription: err.Error()}
				return
			}
			account.BackupDirInfo.Output = resp.Output
			account.BackupDirInfo.DirExists = resp.DirExists
			account.BackupDirInfo.PathExists = resp.PathExists
			account.BackupDirInfo.BackupDir = resp.BackupDir
		})
		if i%batch == 0 || i == len(serverInfo.AccountInfos.Accounts)-1 {
			wg.Wait()
		}
	}
}

// loadHardwareInfo 加载硬件相关信息
func (s *ServersService) loadHardwareInfo(es ExecutorService, arg *internal_models.LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	if !arg.WithHardwareInfo {
		return
	}
	serverInfo.HardwareInfo = &internal_models.ServerHardwareInfo{
		CPUHardwareInfo: &internal_models.ServerCPUHardwareInfo{
			ServerInfoCommon: &internal_models.ServerInfoCommon{
				Output:     "",
				FailedInfo: nil,
			},
			Info: nil,
		},
		GPUHardwareInfos: &internal_models.ServerGPUHardwareInfos{
			ServerInfoCommon: &internal_models.ServerInfoCommon{
				Output:     "",
				FailedInfo: nil,
			},
			Infos: nil,
		},
	}
	// CPU
	cpuResp, err := es.GetCPUHardware()
	serverInfo.HardwareInfo.CPUHardwareInfo.Output = cpuResp.Output
	if err != nil {
		serverInfo.HardwareInfo.CPUHardwareInfo.FailedInfo = &internal_models.ServerInfoLoadingFailedInfo{
			CauseDescription: fmt.Sprintf("向服务器查询cpu数据时出错！es=[%s]，出错信息为：[%s]", es, err.Error()),
		}
	}
	serverInfo.HardwareInfo.CPUHardwareInfo.Info = cpuResp.CPU
	// GPU
	gpuResp, err := es.GetGPUHardware()
	serverInfo.HardwareInfo.GPUHardwareInfos.Output = gpuResp.Output
	if err != nil {
		serverInfo.HardwareInfo.GPUHardwareInfos.FailedInfo = &internal_models.ServerInfoLoadingFailedInfo{
			CauseDescription: fmt.Sprintf("向服务器查询gpu数据时出错！es=[%s]，出错信息为：[%s]", es, err.Error()),
		}
	}
	serverInfo.HardwareInfo.GPUHardwareInfos.Infos = gpuResp.GPUs
}

// loadRemoteAccessUsages 加载正在远程访问该Server的用户使用信息。
func (s *ServersService) loadRemoteAccessUsages(es ExecutorService, arg *internal_models.LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	if !arg.WithRemoteAccessUsages {
		return
	}
	serverInfo.RemoteAccessingUsageInfo = &internal_models.ServerRemoteAccessingUsagesInfo{
		ServerInfoCommon: &internal_models.ServerInfoCommon{
			Output:     "",
			FailedInfo: nil,
		},
		Infos: nil,
	}
	resp, err := es.GetRemoteAccessInfos()
	serverInfo.RemoteAccessingUsageInfo.Output = resp.Output
	if err != nil {
		serverInfo.RemoteAccessingUsageInfo.FailedInfo = &internal_models.ServerInfoLoadingFailedInfo{
			CauseDescription: fmt.Sprintf("向服务器查询远端访问数据时出错！es=[%s]，出错信息为：[%s]", es, err.Error()),
		}
		return
	}
	serverInfo.RemoteAccessingUsageInfo.Infos = resp.RemoteAccessingAccountInfos
}

// loadGPUUsages 加载GPU的使用情况
func (s *ServersService) loadGPUUsages(es ExecutorService, arg *internal_models.LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	if !arg.WithGPUUsages {
		return
	}
	serverInfo.GPUUsageInfo = &internal_models.ServerGPUUsageInfo{
		ServerInfoCommon: &internal_models.ServerInfoCommon{
			Output:     "",
			FailedInfo: nil,
		},
	}
	resp, err := es.GetGPUUsages()
	if err != nil {
		serverInfo.GPUUsageInfo.FailedInfo.CauseDescription = fmt.Sprintf("向服务器查询GPU使用数据时出错！es=[%s]，出错信息为：[%s]", es, err.Error())
		return
	}
	serverInfo.GPUUsageInfo.Output = resp.Output
}

// loadCPUMemProcessesUsageInfo 加载当前正在使用CPU，内存，以及进程的占用信息。
func (s *ServersService) loadCPUMemProcessesUsageInfo(es ExecutorService, arg *internal_models.LoadServerDetailArg, serverInfo *internal_models.ServerInfo) {
	if !arg.WithCPUMemProcessesUsage {
		return
	}
	serverInfo.CPUMemProcessesUsageInfo = &internal_models.ServerCPUMemProcessesUsageInfo{
		ServerInfoCommon: &internal_models.ServerInfoCommon{},
	}
	topResp, err := es.GetCPUMemProcessesUsages()
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

func (s *ServersService) packServer(server *daModels.Server) *internal_models.ServerBasic {
	return &internal_models.ServerBasic{
		CreatedAt:        server.CreatedAt,
		UpdatedAt:        server.UpdatedAt,
		DeletedAt:        server.DeletedAt,
		Host:             server.Host,
		Port:             server.Port,
		AdminAccountName: server.AdminAccountName,
		AdminAccountPwd:  server.AdminAccountPwd,
		OSType:           server.OSType,
	}
}

func (s *ServersService) packAccount(account *daModels.Account) *internal_models.ServerAccount {
	return &internal_models.ServerAccount{
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
		DeletedAt: account.DeletedAt,
		Name:      account.Name,
		Pwd:       account.Pwd,
		Host:      account.Host,
		Port:      account.Port,
	}
}

func (s *ServersService) packPtrAccounts(accounts []*daModels.Account) []*internal_models.ServerAccount {
	res := make([]*internal_models.ServerAccount, 0, len(accounts))
	for _, ac := range accounts {
		res = append(res, s.packAccount(ac))
	}
	return res
}

func (s *ServersService) packAccounts(accounts []daModels.Account) []*internal_models.ServerAccount {
	res := make([]*internal_models.ServerAccount, 0, len(accounts))
	for _, ac := range accounts {
		res = append(res, s.packAccount(&ac))
	}
	return res
}
