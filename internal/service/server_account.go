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
	exists, err := s.accountNameExists(c, Host, Port, AccountName)
	if err != nil {
		return err
	}
	if exists {
		return SErr.CreateAccountNameAlreadyExists
	}
	return s.doAddAccount(Host, Port, AccountName, AccountPwd)
}

func (s *ServersService) accountNameExists(c *gin.Context, Host string, Port uint, AccountName string) (bool, *SErr.APIErr) {
	serverInfo, err := s.Info(c, Host, Port, &internal_models.LoadServerDetailArg{
		WithAccounts:             true,
	})
	if err != nil || serverInfo.AccountInfos.FailedInfo != nil {
		msg := fmt.Sprintf("添加账户前，查询账户列表时出错！出错信息为：err=[%s]，AccountInfos=[%s]，ES=[%s]", err.Error(), util.Pretty(serverInfo.AccountInfos), s.ES)
		log.Println(msg)
		return false, err.CustomMessage(msg)
	}
	for _, account := range serverInfo.AccountInfos.Accounts {
		if account.Name == AccountName {
			return true, nil
		}
	}
	return false, nil
}

func (s *ServersService) doAddAccount(Host string, Port uint, AccountName, AccountPwd string) *SErr.APIErr {
	if s.ES == nil {
		panic(fmt.Sprintf("ServersService doAddAccount s.ES == nil, this method must called with ES inited."))
	}
	resp, err := s.ES.AddAccount(AccountPwd, AccountPwd)
	if err != nil {
		log.Printf("ServersService AddAccount Failed ES=[%s], AccountName=[%s], AccountPwd=[%s], resp=[%s]", s.ES, AccountName, AccountPwd, util.Pretty(resp))
		msg := fmt.Sprintf("添加账户失败！出错信息为：err=[%s]", err.Error())
		log.Println(msg)
		return err.CustomMessage(msg)
	}
	acc := &daModels.Account{
		Name:      AccountName,
		Pwd:       AccountPwd,
		Host:      Host,
		Port:      Port,
	}
	accDal := dal.GetAccountDal()
	err = accDal.Upsert([]*daModels.Account{acc})
	if err != nil {
		// 该MySQL操作失败也没关系，因为这不是关键操作。
		log.Printf("ServersService Upsert Account to MySQL Failed, err=[%+v] ES=[%s], AccountName=[%s], AccountPwd=[%s]", err, s.ES, AccountName, AccountPwd)
	}
	return nil
}


// DeleteAccount 删除账户，可选是否对home目录进行备份，如果需要备份，则返回它备份后的目标文件夹。
func (s *ServersService) DeleteAccount(c *gin.Context, Host string, Port uint, AccountName string, Backup bool) (string, *SErr.APIErr) {
	exists, err := s.accountNameExists(c, Host, Port, AccountName)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", SErr.DeleteAccountNameNotExists
	}
	if s.ES == nil {
		log.Printf("ServersService DeleteAccount s.ES == nil after using s.Info")
		return "", SErr.InternalErr
	}
	var targetBackupDir string
	if Backup {
		// 需要将home目录进行备份。
		resp, err := s.ES.BackupAccountHomeDir(AccountName)
		if err != nil {
			msg := fmt.Sprintf("备份该用户的home目录失败！出错原因：[%s]，ES=[%s]， 请手动检查！", err.Error(), s.ES)
			log.Println(msg)
			return "", err.CustomMessage(msg)
		}
		targetBackupDir = resp.TargetDir
	}
	_, err = s.ES.DeleteAccount(AccountName)
	if err != nil {
		msg := fmt.Sprintf("删除账户失败！出错原因：[%s]，ES=[%s]", err.Error(), s.ES)
		log.Println(msg)
		return "", err.CustomMessage(msg)
	}
	return targetBackupDir, nil
}

// RecoverAccount 恢复某个被删除的账户，可选是否对backup的home目录进行恢复。
func (s *ServersService) RecoverAccount(c *gin.Context, Host string, Port uint, AccountName, AccountPwd string, RecoverBackup bool) *SErr.APIErr {
	exists, err := s.accountNameExists(c, Host, Port, AccountName)
	if err != nil {
		return err
	}
	if exists {
		return SErr.CreateAccountNameAlreadyExists
	}
	if s.ES == nil {
		log.Printf("ServersService RecoverAccount s.ES == nil after using s.Info")
		return SErr.InternalErr
	}
	if RecoverBackup {
		_, err := s.ES.RecoverAccountHomeDir(AccountName, false)
		if err != nil {
			msg := fmt.Sprintf("恢复账户home目录失败！出错原因：[%s]，ES=[%s]", err.Error(), s.ES)
			log.Println(msg)
			return err.CustomMessage(msg)
		}
	}
	return s.doAddAccount(Host, Port, AccountName, AccountPwd)
}

// UpdateAccount 更新某个存在的账户，更新它在MySQL中的存储。这里可能是在Server中存在，但是在MySQL中存储的数据不全时需要使用。
func (s *ServersService) UpdateAccount(c *gin.Context, Host string, Port uint, AccountName, AccountPwd string) *SErr.APIErr {
	exists, err := s.accountNameExists(c, Host, Port, AccountName)
	if err != nil {
		return err
	}
	if !exists {
		return SErr.UpdateAccountNameNotExists
	}
	acc := &daModels.Account{
		Name:      AccountName,
		Pwd:       AccountPwd,
		Host:      Host,
		Port:      Port,
	}
	accDal := dal.GetAccountDal()
	err = accDal.Upsert([]*daModels.Account{acc})
	if err != nil {
		log.Printf("ServersService Upsert Account to MySQL Failed, err=[%+v] ES=[%s], AccountName=[%s], AccountPwd=[%s]", err, s.ES, AccountName, AccountPwd)
		return err
	}
	return nil
}


// BackupDirInfo 查看该用户名对应的backup文件夹的信息。
func (s *ServersService) BackupDirInfo(c *gin.Context, Host string, Port uint, AccountName string) (*internal_models.ServerAccountBackupDirInfo, *SErr.APIErr) {
	_, err := s.Info(c, Host, Port, &internal_models.LoadServerDetailArg{})
	if err != nil {
		return nil, err
	}
	if s.ES == nil {
		log.Printf("ServersService BackupDirInfo s.ES == nil after using s.Info")
		return nil, SErr.InternalErr
	}
	backupDirResp, err := s.ES.GetBackupDir(AccountName)
	if err != nil {
		return nil, err
	}
	return &internal_models.ServerAccountBackupDirInfo{
		BackupDir:  backupDirResp.BackupDir,
		PathExists: backupDirResp.PathExists,
		DirExists:  backupDirResp.DirExists,
	}, nil
}
