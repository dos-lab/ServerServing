package internal_models

type ServerAccountCreateRequest struct {
	Host        string `json:"host"`
	Port        uint   `json:"port"`
	AccountName string `json:"account_name"`
	AccountPwd  string `json:"account_pwd"`
}

type ServerAccountCreateResponse struct {
}

type ServerAccountDeleteRequest struct {
	Host        string `json:"host"`
	Port        uint   `json:"port"`
	AccountName string `json:"account_name"`
	Backup      bool   `json:"backup"`
}

type ServerAccountDeleteResponse struct {
	BackupDir string `json:"backup_dir"`
}

type ServerAccountUpdateRequest struct {
	Recover       bool `form:"recover"`        // Recover 该账户是从删除的账户中恢复
	RecoverBackup bool `form:"recover_backup"` // RecoverBackup 指定是否要恢复backup的用户目录文件夹。

	Host        string `json:"host" form:"host"`
	Port        uint   `json:"port" form:"port"`
	AccountName string `json:"account_name" form:"account_name"`
	AccountPwd  string `json:"account_pwd" form:"account_pwd"`
}

type ServerAccountUpdateResponse struct {
}

type ServerAccountBackupDirRequest struct {
	Host        string `json:"host" form:"host"`
	Port        uint   `json:"port" form:"port"`
	AccountName string `json:"account_name" form:"account_name"`
}

type ServerAccountBackupDirResponse struct {
	ServerAccountBackupDirInfo
}

type ServerAccountBackupDirInfo struct {
	BackupDir  string `json:"backup_dir"`
	PathExists bool   `json:"path_exists"`
	DirExists  bool   `json:"dir_exists"`
}
