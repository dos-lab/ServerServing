package server_executor

import (
	daModels "ServerServing/da/mysql/da_models"
	SErr "ServerServing/err"
	"ServerServing/internal/internal_models"
	"fmt"
	"github.com/tredoe/osutil/v2/userutil/crypt/sha512_crypt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

type LinuxOSType string

const (
	Unknown LinuxOSType = "unknown"
	Ubuntu  LinuxOSType = "ubuntu"
	CentOS  LinuxOSType = "centos"
)

type OpenExecutorServiceParam struct {
	Host             string
	Port             uint
	OSType           daModels.OSType
	AdminAccountName string
	AdminAccountPwd  string
}

func OpenExecutorService(param *OpenExecutorServiceParam) (ExecutorService, *SErr.APIErr) {
	switch param.OSType {
	case daModels.OSTypeLinux:
		return openLinuxSSHExecutorService(param.Host, param.Port, param.AdminAccountName, param.AdminAccountPwd)
	default:
		panic("Unimplemented")
	}
}

// use loadCmdScript to load a file-base command or script
var loadCmdScript = func() func(path, name string) (string, *SErr.APIErr) {
	cache := map[string]string{}
	refresh := map[string]time.Time{}
	refreshInterval := 1 * time.Minute
	genKey := func(path, name string) string {
		return fmt.Sprintf("Path:%s, Name:%s", path, name)
	}
	return func(fPath, name string) (string, *SErr.APIErr) {
		cmdKey := genKey(fPath, name)
		if _, ok := refresh[cmdKey]; !ok {
			refresh[cmdKey] = time.Now()
		}
		if lastRefresh, ok := refresh[cmdKey]; ok && time.Now().Sub(lastRefresh) < refreshInterval {
			cmd := cache[cmdKey]
			if cmd != "" {
				return cmd, nil
			}
		}
		p := path.Join(fPath, name)
		f, err := os.Open(p)
		if err != nil {
			panic(err)
		}
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			log.Printf("loadCmdScript ioutil.readAll failed")
			return "", SErr.InternalErr.CustomMessageF("loadCmdScript加载命令行数据失败！err=[%s]", err.Error())
		}
		cmd := string(bs)
		cache[cmdKey] = cmd
		return cmd, nil
	}
}()

type ExecutorFileSystemService interface {
	Move(src, dst string, force bool) (*ExecutorServiceVoidResp, *SErr.APIErr)
	FileExists(filepath string) (*ExecutorServiceExistsResp, *SErr.APIErr)
	DirExists(dirPath string) (*ExecutorServiceExistsResp, *SErr.APIErr)
	PathExists(path string) (*ExecutorServiceExistsResp, *SErr.APIErr)
	Mkdir(dirPath string) (*ExecutorServiceVoidResp, *SErr.APIErr)
	MkdirIfNotExists(dirPath string) (*ExecutorServiceVoidResp, *SErr.APIErr)
}

type ExecutorHardwareUsageService interface {
	GetCPUMemProcessesUsages() (*ExecutorServiceCPUMemProcessesUsagesResp, *SErr.APIErr)
	GetGPUUsages() (*ExecutorServiceVoidResp, *SErr.APIErr)
}

type ExecutorAccountService interface {
	AddAccount(accountName, pwd string) (*ExecutorServiceVoidResp, *SErr.APIErr)
	DeleteAccount(accountName string) (*ExecutorServiceVoidResp, *SErr.APIErr)
	GetAccountList() (*ExecutorServiceGetAccountListResp, *SErr.APIErr)
	GetBackupDir(accountName string) (*ExecutorServiceGetBackupDirResp, *SErr.APIErr)
	BackupAccountHomeDir(accountName string) (*ExecutorServiceBackupAccountResp, *SErr.APIErr)
	RecoverAccountHomeDir(accountName string, force bool) (*ExecutorServiceRecoverAccountResp, *SErr.APIErr)
	GetAccountHomeDir(accountName string) (*ExecutorServiceGetAccountHomeDirResp, *SErr.APIErr)
}

type ExecutorHardwareInfoService interface {
	GetGPUHardware() (*ExecutorServiceGPUHardwareResp, *SErr.APIErr)
	GetCPUHardware() (*ExecutorServiceCPUHardwareResp, *SErr.APIErr)
	GetMemoryHardware() (*ExecutorServiceMemoryHardwareResp, *SErr.APIErr)
}

type ExecutorRemoteAccessService interface {
	GetRemoteAccessInfos() (*ExecutorServiceRemoteAccessResp, *SErr.APIErr)
}

// ExecutorService 描述远端命令组成的的外部可用接口。目前只包括Linux服务器。
// 其中每个接口的第一个返回参数永远都是从服务器返回的真实output，用于在复杂情况下debug，或者直接给用户展示它的内容。
type ExecutorService interface {
	ExecutorFileSystemService
	ExecutorHardwareUsageService
	ExecutorAccountService
	ExecutorHardwareInfoService
	ExecutorRemoteAccessService
	io.Closer
	String() string
}

type ExecutorServiceRespCommon struct {
	Output string
}

type ExecutorServiceVoidResp struct {
	ExecutorServiceRespCommon
}

type ExecutorServiceExistsResp struct {
	ExecutorServiceRespCommon
	Exists bool
}

type ExecutorServiceBoolResp struct {
	ExecutorServiceRespCommon
	Result bool
}

type ExecutorServiceBackupAccountResp struct {
	ExecutorServiceRespCommon
	TargetDir string
}

type ExecutorServiceRecoverAccountResp struct {
	ExecutorServiceRespCommon
	HomeDir string
}

type ExecutorServiceGetAccountHomeDirResp struct {
	ExecutorServiceRespCommon
	HomeDir string
}

type ExecutorServiceGetAccountListResp struct {
	ExecutorServiceRespCommon
	Accounts []*internal_models.ServerAccount
}

type ExecutorServiceCPUMemProcessesUsagesResp struct {
	ExecutorServiceRespCommon
	CPUMemUsage  *internal_models.ServerCPUMemUsage
	ProcessInfos []*internal_models.ServerProcessInfo
}

type ExecutorServiceCPUHardwareResp struct {
	ExecutorServiceRespCommon
	CPU *internal_models.ServerCPUs
}

type ExecutorServiceGPUHardwareResp struct {
	ExecutorServiceRespCommon
	GPUs []*internal_models.ServerGPU
}

type ExecutorServiceMemoryHardwareResp struct {
	ExecutorServiceRespCommon
	MemoryStats *internal_models.ServerMemory
}

type ExecutorServiceRemoteAccessResp struct {
	ExecutorServiceRespCommon
	RemoteAccessingAccountInfos []*internal_models.ServerRemoteAccessingAccount
}

type ExecutorServiceGetBackupDirResp struct {
	ExecutorServiceRespCommon
	BackupDir  string
	PathExists bool
	DirExists  bool
}

type executorServiceCommon struct{}

func (c *executorServiceCommon) encrypt(pwd string) string {
	// Generate a random string for use in the salt
	//const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	//seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	//s := make([]byte, 8)
	//for i := range s {
	//	s[i] = charset[seededRand.Intn(len(charset))]
	//}
	//salt := []byte(fmt.Sprintf("$6$%s", s))
	salt := []byte("$6$salt")
	// use salt to hash user-supplied password
	sc := sha512_crypt.New()
	hash, err := sc.Generate([]byte(pwd), salt)
	if err != nil {
		log.Printf("error hashing password %v", err)
		panic("error hashing user's supplied password: %s\n")
	}
	return hash
}
