package server_executor

import (
	"ServerServing/config"
	"ServerServing/util"
	"testing"
)

var es ExecutorService

func initEnv(t *testing.T) {
	if es != nil {
		return
	}
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	//s, err := openLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	s, err := openLinuxSSHExecutorService("114.116.101.120", 22, "someadmin", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	es = s
}

func TestSSHExecutorService_GetAccountList(t *testing.T) {
	initEnv(t)
	resp, e := es.GetAccountList()
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("accountList=[%v]", util.Pretty(resp))
}

func TestSSHExecutorService_CreateAccount(t *testing.T) {
	initEnv(t)
	_, e := es.AddAccount("golang_test_acc_1", "123456")
	if e != nil {
		t.Fatal(e)
	}
}

func TestSSHExecutorService_DeleteAccount(t *testing.T) {
	initEnv(t)
	resp, e := es.DeleteAccount("golang_test_acc_1")
	t.Log(resp)
	if e != nil {
		t.Fatal(e)
	}
}

func TestSSHExecutorService_Path(t *testing.T) {
	initEnv(t)
	resp, e := es.PathExists("/root")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = es.DirExists("/")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = es.DirExists("/home/mynewuser/some_file.txt")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
}

func TestSSHExecutorService_FileSystem(t *testing.T) {
	initEnv(t)
	output, e := es.Mkdir("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output)
	resp, e := es.PathExists("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = es.PathExists("/dummy")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = es.DirExists("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = es.FileExists("/home/mynewuser/some_file.txt")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	mvResp, e := es.Move("/some_dir", "/some_dir_2", false)
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(mvResp))
}

func TestSSHExecutorService_Backup(t *testing.T) {
	initEnv(t)
	resp, err := es.BackupAccountHomeDir("golang_test_acc_1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestSSHExecutorService_Recover(t *testing.T) {
	initEnv(t)
	resp, err := es.RecoverAccountHomeDir("golang_test_acc_1", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestSSHExecutorService_Top(t *testing.T) {
	initEnv(t)
	resp, err := es.GetCPUMemProcessesUsages()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestConnect(t *testing.T) {
	initEnv(t)
	// c, err := openLinuxSSHConnection("47.93.56.75", 22, "someuser", "zhjT9910123!")
	// c, err := openLinuxSSHConnection("47.93.56.75:22", "mynewuser", "123456")
	c, err := openLinuxSSHConnection("114.116.101.120", 22, "someadmin", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	output, err := c.SendCommands("top -bn1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(output))
}

func TestGPUInfo(t *testing.T) {
	initEnv(t)
	resp, err := es.GetGPUHardware()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestCPUInfo(t *testing.T) {
	initEnv(t)
	resp, err := es.GetCPUHardware()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestRemoteAccess(t *testing.T) {
	initEnv(t)
	resp, err := es.GetRemoteAccessInfos()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestMemInfo(t *testing.T) {
	initEnv(t)
	resp, err := es.GetMemoryHardware()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestGPUUsage(t *testing.T) {
	initEnv(t)
	resp, err := es.GetGPUUsages()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("output:", resp.Output)
	t.Log(util.Pretty(resp))
}
