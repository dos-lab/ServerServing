package internal_models

type Server struct {
	Host             string `json:"host"`
	Port             uint   `json:"port"` // ssh port
	AdminAccountName string `json:"admin_account_name"`
	AdminAccountPwd  string `json:"admin_account_pwd"`
}

type Account struct {
	Name string `json:"name"`
	Pwd  string `json:"pwd"`
	UID  int    `json:"uid"`
	GID  int    `json:"gid"`
}
