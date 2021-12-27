package service

import (
	"ServerServing/config"
	"ServerServing/util"
	"fmt"
	"github.com/melbahja/goph"
	"log"
	"testing"
)

func TestSSHExecutorService_GetAccountList(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := OpenLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	resp, e := s.GetAccountList()
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("accountList=[%v]", util.Pretty(resp))
}

func TestSSHExecutorService_CreateAccount(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := OpenLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	_, e := s.AddAccount("user_created_by_golang_1", "123456")
	if e != nil {
		t.Fatal(e)
	}
}

func TestSSHExecutorService_DeleteAccount(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	// s, err := OpenLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	s, err := OpenLinuxSSHExecutorService("133.133.135.42", 22, "yzc", "zhjt9910")
	if err != nil {
		t.Fatal(err)
	}
	resp, e := s.DeleteAccount("yzc_test")
	t.Log(resp)
	if e != nil {
		t.Fatal(e)
	}
}

func TestSSHExecutorService_Path(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := OpenLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	resp, e := s.PathExists("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = s.DirExists("/dummy")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = s.DirExists("/home/mynewuser/some_file.txt")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
}

func TestSSHExecutorService_FileSystem(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := OpenLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	output, e := s.Mkdir("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output)
	resp, e := s.PathExists("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = s.PathExists("/dummy")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = s.DirExists("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	resp, e = s.FileExists("/home/mynewuser/some_file.txt")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(resp))
	mvResp, e := s.Move("/some_dir", "/some_dir_2", false)
	if e != nil {
		t.Fatal(e)
	}
	t.Log(util.Pretty(mvResp))
}

func TestSSHExecutorService_Backup(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := OpenLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := s.BackupAccountHomeDir("mynewuser")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestSSHExecutorService_Recover(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := OpenLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := s.RecoverAccountHomeDir("mynewuser", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestSSHExecutorService_Top(t *testing.T) {
	s := initConn(t)
	resp, err := s.GetCPUMemProcessesUsages()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func initConn(t *testing.T) ExecutorService {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := OpenLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestConnect(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	c, err := ConnectLinuxSSH("47.93.56.75", 22, "someuser", "zhjT9910123!")
	// c, err := ConnectLinuxSSH("47.93.56.75:22", "mynewuser", "123456")
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
	s := initConn(t)
	resp, err := s.GetGPUHardware()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestCPUInfo(t *testing.T) {
	s := initConn(t)
	resp, err := s.GetCPUHardware()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestRemoteAccess(t *testing.T) {
	s := initConn(t)
	resp, err := s.GetRemoteAccessInfos()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestMemInfo(t *testing.T) {
	s := initConn(t)
	resp, err := s.GetMemoryHardware()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(util.Pretty(resp))
}

func TestGPUUsage(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := OpenLinuxSSHExecutorService("133.133.135.42", 22, "yzc", "zhjt9910")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := s.GetGPUUsages()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("output:", resp.Output)
	t.Log(util.Pretty(resp))
}

func TestGorh(t *testing.T) {
	// Start new ssh connection with private key.
	auth := goph.Password("zhjT9910123!")

	client, err := goph.New("someuser", "47.93.56.75", auth)
	if err != nil {
		log.Fatal(err)
	}

	// Defer closing the network connection.
	defer client.Close()

	// Execute your command.
	out, err := client.Run("top -bn1")

	if err != nil {
		log.Fatal(err)
	}

	// Get your output as []byte.
	fmt.Println(string(out))
}
