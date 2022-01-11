package service

import (
	"ServerServing/config"
	"ServerServing/da/mysql"
	"ServerServing/internal/internal_models"
	"ServerServing/util"
	"testing"
)

func initEnv(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	mysql.InitMySQL()
}

func TestServersService_Info(t *testing.T) {
	initEnv(t)
	svc := GetServersService()
	res, err := svc.Info(nil, "114.116.101.120", 22, &internal_models.LoadServerDetailArg{
		WithHardwareInfo:             true,
		WithAccounts:                 true,
		WithAccountsIgnoreDBAccounts: false,
		WithRemoteAccessUsages:       true,
		WithGPUUsages:                true,
		WithCPUMemProcessesUsage:     true,
		WithBackupDirInfo:            false,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(res))
}
