package mysql

import (
	"ServerServing/config"
	"ServerServing/da/mysql/da_models"
	"ServerServing/util"
	"strconv"
	"testing"
)

func TestInitMySQL(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	InitMySQL()
	selectServerAndAccounts(t)
	// initServer(t)
	// initAccounts()
}

func initServer(t *testing.T) {
	s := &da_models.Server{
		Host:             "47.93.56.75",
		Port:             22,
		AdminAccountName: "root",
		AdminAccountPwd:  "zhjT9910123!",
	}
	res := db.Model(&da_models.Server{}).Create(s)
	if res.Error != nil {
		t.Fatal(res.Error)
	}
	t.Log(util.Pretty(s))
}

func initAccounts() {
	accs := make([]*da_models.Account, 0, 10)
	for i := 0; i < 10; i++ {
		accs = append(accs, &da_models.Account{
			Name: strconv.FormatInt(int64(i), 10),
			Pwd:  "123456",
			Server: da_models.Server{
				Host: "47.93.56.75",
				Port: 22,
			},
		})
	}
	db.Model(&da_models.Account{}).Create(accs)
}

func selectServerAndAccounts(t *testing.T) {
	var s []*da_models.Server
	res := db.Model(&da_models.Server{}).Preload("Accounts").Where("host = ? and port = ?", "47.93.56.75", 22).Find(&s)
	if res.Error != nil {
		t.Fatal(res.Error)
		return
	}
	t.Log(util.Pretty(s))
}

func selectAccounts(t *testing.T) {
	var accs []*da_models.Account
	res := db.Model(&da_models.Account{}).Joins("Server").Find(&accs)
	if res.Error != nil {
		t.Fatal(res.Error)
		return
	}
	t.Log(util.Pretty(accs))
}
