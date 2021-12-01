package service

import (
	"ServerServing/config"
	"ServerServing/util"
	"testing"
)

func TestSSHExecutorService_GetAccountList(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := GetLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	_, accountList, e := s.GetAccountList()
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("accountList=[%v]", util.Pretty(accountList))
}

func TestSSHExecutorService_CreateAccount(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := GetLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	_, e := s.AddAccount("user_created_by_golang", "123456")
	if e != nil {
		t.Fatal(e)
	}
}

func TestSSHExecutorService_DeleteAccount(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := GetLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	_, e := s.DeleteAccount("user_created_by_golang")
	if e != nil {
		t.Fatal(e)
	}
}

func TestSSHExecutorService_Path(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := GetLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	output, exists, e := s.PathExists("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output, exists)
	output, exists, e = s.DirExists("/dummy")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output, exists)
	output, exists, e = s.DirExists("/home/mynewuser/some_file.txt")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output, exists)
}

func TestSSHExecutorService_FileSystem(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := GetLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	output, e := s.Mkdir("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output)
	output, exists, e := s.PathExists("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output, exists)
	output, exists, e = s.PathExists("/dummy")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output, exists)
	output, exists, e = s.DirExists("/some_dir")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output, exists)
	output, exists, e = s.FileExists("/home/mynewuser/some_file.txt")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output, exists)
	output, e = s.Move("/some_dir", "/some_dir_2", false)
	if e != nil {
		t.Fatal(e)
	}
	t.Log(output, exists)
}

func TestSSHExecutorService_Backup(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := GetLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	output, targetDir, err := s.BackupAccountHomeDir("mynewuser")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(output, targetDir)
}

func TestSSHExecutorService_Recover(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	s, err := GetLinuxSSHExecutorService("47.93.56.75", 22, "someuser", "zhjT9910123!")
	if err != nil {
		t.Fatal(err)
	}
	output, targetDir, err := s.RecoverAccountHomeDir("mynewuser", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(output, targetDir)
}

func TestConnect(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	c, err := ConnectLinuxSSH("47.93.56.75:22", "someuser", "zhjT9910123!")
	// c, err := ConnectLinuxSSH("47.93.56.75:22", "mynewuser", "123456")
	if err != nil {
		t.Fatal(err)
	}
	output, err := c.SendCommands("sudo cat /etc/os-release")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(output))
}
