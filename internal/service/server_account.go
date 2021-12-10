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
)

func (s *ServersService) AddAccount(c *gin.Context, Host string, Port uint, AccountName, AccountPwd string) *SErr.APIErr {
	es, err := s.OpenExecutorServiceByHostPort(c, Host, Port)
	if err != nil {
		msg := fmt.Sprintf("添加账户时，打开连接时出错！出错信息为：err=[%s], Host=[%s], Port=[%d], AccountName=[%s]", err.Error(), Host, Port, AccountName)
		log.Println(msg)
		return err.CustomMessage(msg)
	}
	defer func() {
		_ = es.Close()
	}()
	exists, err := s.accountNameExists(c, es, Host, Port, AccountName)
	if err != nil {
		return err
	}
	if exists {
		return SErr.CreateAccountNameAlreadyExists
	}
	return s.doAddAccount(c, es, Host, Port, AccountName, AccountPwd)
}

func (s *ServersService) accountNameExists(c *gin.Context, es ExecutorService, Host string, Port uint, AccountName string) (bool, *SErr.APIErr) {
	serverInfo := &internal_models.ServerInfo{}
	s.loadInfoFromServer(serverInfo, es, &internal_models.LoadServerDetailArg{
		WithAccounts: true,
	})
	if serverInfo.AccountInfos.FailedInfo != nil {
		msg := fmt.Sprintf("查询账户列表时出错！出错信息为：AccountInfos=[%s], Host=[%s], Port=[%d], AccountName=[%s]", util.Pretty(serverInfo.AccountInfos), Host, Port, AccountName)
		log.Println(msg)
		return false, SErr.InternalErr.CustomMessage(msg)
	}
	for _, account := range serverInfo.AccountInfos.Accounts {
		if account.Name == AccountName {
			return true, nil
		}
	}
	return false, nil
}

func (s *ServersService) doAddAccount(c *gin.Context, es ExecutorService, Host string, Port uint, AccountName, AccountPwd string) *SErr.APIErr {
	resp, err := es.AddAccount(AccountName, AccountPwd)
	if err != nil {
		log.Printf("ServersService AddAccount Failed ES=[%s], AccountName=[%s], AccountPwd=[%s], resp=[%s]", es, AccountName, AccountPwd, util.Pretty(resp))
		msg := fmt.Sprintf("添加账户失败！出错信息为：err=[%s]", err.Error())
		log.Println(msg)
		return err.CustomMessage(msg)
	}
	acc := &daModels.Account{
		Name: AccountName,
		Pwd:  AccountPwd,
		Host: Host,
		Port: Port,
	}
	accDal := dal.GetAccountDal()
	err = accDal.Upsert([]*daModels.Account{acc})
	if err != nil {
		// 该MySQL操作失败也没关系，因为这不是关键操作。
		log.Printf("ServersService Upsert Account to MySQL Failed, err=[%+v] ES=[%s], AccountName=[%s], AccountPwd=[%s]", err, es, AccountName, AccountPwd)
	}
	return nil
}

// DeleteAccount 删除账户，可选是否对home目录进行备份，如果需要备份，则返回它备份后的目标文件夹。
func (s *ServersService) DeleteAccount(c *gin.Context, Host string, Port uint, AccountName string, Backup bool) (string, *SErr.APIErr) {
	es, err := s.OpenExecutorServiceByHostPort(c, Host, Port)
	if err != nil {
		msg := fmt.Sprintf("删除账户时，打开连接时出错！出错信息为：err=[%s], Host=[%s], Port=[%d], AccountName=[%s]", err.Error(), Host, Port, AccountName)
		log.Println(msg)
		return "", err.CustomMessage(msg)
	}
	defer func() {
		_ = es.Close()
	}()
	exists, err := s.accountNameExists(c, es, Host, Port, AccountName)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", SErr.DeleteAccountNameNotExists
	}
	var targetBackupDir string
	if Backup {
		// 需要将home目录进行备份。
		resp, err := es.BackupAccountHomeDir(AccountName)
		if err != nil {
			msg := fmt.Sprintf("备份该用户的home目录失败！出错原因：[%s]，ES=[%s]， 请手动检查！", err.Error(), es)
			log.Println(msg)
			return "", err.CustomMessage(msg)
		}
		targetBackupDir = resp.TargetDir
	}
	_, err = es.DeleteAccount(AccountName)
	if err != nil {
		msg := fmt.Sprintf("删除账户失败！出错原因：[%s]，ES=[%s]", err.Error(), es)
		log.Println(msg)
		return "", err.CustomMessage(msg)
	}
	return targetBackupDir, nil
}

// RecoverAccount 恢复某个被删除的账户，可选是否对backup的home目录进行恢复。
func (s *ServersService) RecoverAccount(c *gin.Context, Host string, Port uint, AccountName, AccountPwd string, RecoverBackup bool) *SErr.APIErr {
	es, err := s.OpenExecutorServiceByHostPort(c, Host, Port)
	if err != nil {
		msg := fmt.Sprintf("恢复账户时，打开连接时出错！出错信息为：err=[%s], Host=[%s], Port=[%d], AccountName=[%s]", err.Error(), Host, Port, AccountName)
		log.Println(msg)
		return err.CustomMessage(msg)
	}
	defer func() {
		_ = es.Close()
	}()
	exists, err := s.accountNameExists(c, es, Host, Port, AccountName)
	if err != nil {
		return err
	}
	if exists {
		return SErr.CreateAccountNameAlreadyExists
	}
	err = s.doAddAccount(c, es, Host, Port, AccountName, AccountPwd)
	if err != nil {
		msg := fmt.Sprintf("恢复账户时，添加账户时出错！出错信息为：err=[%s], Host=[%s], Port=[%d], AccountName=[%s]", err.Error(), Host, Port, AccountName)
		log.Println(msg)
		return err.CustomMessage(msg)
	}
	if RecoverBackup {
		_, err := es.RecoverAccountHomeDir(AccountName, true)
		if err != nil {
			msg := fmt.Sprintf("恢复账户home目录失败！出错原因：[%s]，ES=[%s]", err.Error(), es)
			log.Println(msg)
			return err.CustomMessage(msg)
		}
	}
	return nil
}

// UpdateAccount 更新某个存在的账户，更新它在MySQL中的存储。这里可能是在Server中存在，但是在MySQL中存储的数据不全时需要使用。
func (s *ServersService) UpdateAccount(c *gin.Context, Host string, Port uint, AccountName, AccountPwd string) *SErr.APIErr {
	es, err := s.OpenExecutorServiceByHostPort(c, Host, Port)
	if err != nil {
		msg := fmt.Sprintf("更新账户时，打开连接时出错！出错信息为：err=[%s], Host=[%s], Port=[%d], AccountName=[%s]", err.Error(), Host, Port, AccountName)
		log.Println(msg)
		return err.CustomMessage(msg)
	}
	defer func() {
		_ = es.Close()
	}()
	exists, err := s.accountNameExists(c, es, Host, Port, AccountName)
	if err != nil {
		return err
	}
	if !exists {
		return SErr.UpdateAccountNameNotExists
	}
	acc := &daModels.Account{
		Name: AccountName,
		Pwd:  AccountPwd,
		Host: Host,
		Port: Port,
	}
	accDal := dal.GetAccountDal()
	err = accDal.Upsert([]*daModels.Account{acc})
	if err != nil {
		log.Printf("ServersService Upsert Account to MySQL Failed, err=[%+v] ES=[%s], AccountName=[%s], AccountPwd=[%s]", err, es, AccountName, AccountPwd)
		return err
	}
	return nil
}

// BackupDirInfo 查看该用户名对应的backup文件夹的信息。
func (s *ServersService) BackupDirInfo(c *gin.Context, Host string, Port uint, AccountName string) (*internal_models.ServerAccountBackupDirInfo, *SErr.APIErr) {
	es, err := s.OpenExecutorServiceByHostPort(c, Host, Port)
	if err != nil {
		msg := fmt.Sprintf("备份账户时，打开连接时出错！出错信息为：err=[%s], Host=[%s], Port=[%d], AccountName=[%s]", err.Error(), Host, Port, AccountName)
		log.Println(msg)
		return nil, err.CustomMessage(msg)
	}
	defer func() {
		_ = es.Close()
	}()
	backupDirResp, err := es.GetBackupDir(AccountName)
	if err != nil {
		return nil, err
	}
	return &internal_models.ServerAccountBackupDirInfo{
		BackupDir:  backupDirResp.BackupDir,
		PathExists: backupDirResp.PathExists,
		DirExists:  backupDirResp.DirExists,
	}, nil
}
