package service

import (
	"ServerServing/config"
	SErr "ServerServing/err"
	"ServerServing/internal/internal_models"
	"ServerServing/util"
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LinuxOSType string

const (
	Unknown LinuxOSType = "unknown"
	Ubuntu  LinuxOSType = "ubuntu"
	CentOS  LinuxOSType = "centos"
)

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
		}
		cmd := string(bs)
		cache[cmdKey] = cmd
		return cmd, nil
	}
}()

// ExecutorService 描述远端命令组成的的外部可用接口。目前只包括Linux服务器。
// 其中每个接口的第一个返回参数永远都是从服务器返回的真实output，用于在复杂情况下debug，或者直接给用户展示它的内容。
type ExecutorService interface {
	Move(src, dst string, force bool) (*ExecutorServiceVoidResp, *SErr.APIErr)
	FileExists(filepath string) (*ExecutorServiceExistsResp, *SErr.APIErr)
	DirExists(dirPath string) (*ExecutorServiceExistsResp, *SErr.APIErr)
	PathExists(path string) (*ExecutorServiceExistsResp, *SErr.APIErr)
	Mkdir(dirPath string) (*ExecutorServiceVoidResp, *SErr.APIErr)
	MkdirIfNotExists(dirPath string) (*ExecutorServiceVoidResp, *SErr.APIErr)

	Top() (*ExecutorServiceTopResp, *SErr.APIErr)

	AddAccount(accountName, pwd string) (*ExecutorServiceVoidResp, *SErr.APIErr)
	DeleteAccount(accountName string) (*ExecutorServiceVoidResp, *SErr.APIErr)
	GetAccountList() (*ExecutorServiceGetAccountListResp, *SErr.APIErr)
	BackupAccountHomeDir(accountName string) (*ExecutorServiceBackupAccountResp, *SErr.APIErr)
	RecoverAccountHomeDir(accountName string, force bool) (*ExecutorServiceRecoverAccountResp, *SErr.APIErr)
	GetAccountHomeDir(accountName string) (*ExecutorServiceGetAccountHomeDirResp, *SErr.APIErr)
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
	Accounts []*internal_models.Account
}

type ExecutorServiceTopResp struct {
	ExecutorServiceRespCommon
	CPUMemUsage *internal_models.ServerCPUMemUsage
	ProcessInfos []*internal_models.ServerProcessInfo
}

// GetLinuxSSHExecutorService
// 获取一个到某服务器的SSH Executor服务实例。建立新的ssh连接。在一次http请求中复用这个服务可以减少连接的创建数量。
// 目前不确定是否可以在多个请求间共享，暂时不太建议这么做（需要测试）。这容易引起并发问题，反正总的并发量不高，目前先随便做一做。
func GetLinuxSSHExecutorService(host string, port uint, account, pwd string) (ExecutorService, *SErr.APIErr) {
	SSHConn, err := ConnectLinuxSSH(host, port, account, pwd)
	if err != nil {
		return nil, SErr.SSHConnectionErr.CustomMessageF("与该服务器建立SSH连接失败！服务器地址为%s:%d，用户名为：%s，密码为：%s", host, port, account, pwd)
	}
	output, hasSudo, err := SSHConn.CheckSudoPrivilege()
	log.Printf("CheckSudoPrivilege output=[%s]", output)
	if err != nil {
		return nil, SErr.SSHConnectionErr.CustomMessageF("检查用户是否具有sudo权限时失败！服务器地址为%s:%d，用户名为：%s，密码为：%s", host, port, account, pwd)
	}
	if !hasSudo {
		return nil, SErr.SSHConnectionErr.CustomMessageF("该用户并不具有sudo权限！服务器地址为%s:%d，用户名为：%s，密码为：%s", host, port, account, pwd)
	}
	output, osType, err := SSHConn.CheckOSInfo()
	log.Printf("CheckOSInfo output=[%s]", output)
	if err != nil {
		return nil, SErr.SSHConnectionErr.CustomMessageF("检查该服务器的操作系统类型失败！服务器地址为%s:%d，用户名为：%s，密码为：%s", host, port, account, pwd)
	}
	if osType == Unknown {
		return nil, SErr.SSHConnectionErr.CustomMessageF("该服务器的操作系统类型为不支持的类型！服务器地址为%s:%d，用户名为：%s，密码为：%s", host, port, account, pwd)
	}
	switch osType {
	case Ubuntu:
		return NewLinuxSSHExecutorServiceCommon(Ubuntu, account, pwd, SSHConn), nil
	case CentOS:
		return NewLinuxSSHExecutorServiceCommon(CentOS, account, pwd, SSHConn), nil
	default:
		panic("Unsupported LinuxOSType")
	}
}

type LinuxSSHExecutorServiceCommon struct {
	Host    string
	Port    uint
	Account string
	Pwd     string

	OSType  LinuxOSType
	SSHConn *LinuxSSHConnection
	commonPath string
	ubuntuPath string
	centosPath string

	ubuntuService *UbuntuSSHExecutorService
}

func NewLinuxSSHExecutorServiceCommon(osType LinuxOSType, Account string, Pwd string, SSHConn *LinuxSSHConnection) *LinuxSSHExecutorServiceCommon {
	csp := config.GetConfig().CmdsScriptsPath
	ubuntu := path.Join(csp, "ubuntu")
	common := path.Join(csp, "common")
	centos := path.Join(csp, "centos")
	return &LinuxSSHExecutorServiceCommon{
		Account:       Account,
		Pwd:           Pwd,
		OSType:        osType,
		SSHConn:       SSHConn,
		commonPath:    common,
		ubuntuPath:    ubuntu,
		centosPath:    centos,
		ubuntuService: NewUbuntuSSHExecutorService(SSHConn),
	}

}

func (s *LinuxSSHExecutorServiceCommon) String() string {
	return fmt.Sprintf("LinuxSSHExecutorServiceCommon=[Host=%s, Port=%d, Account=%s, Pwd=%s]", s.Host, s.Port, s.Account, s.Pwd)
}

// Move 移动文件或文件夹
func (s *LinuxSSHExecutorServiceCommon) Move(src, dst string, force bool) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	resp := &ExecutorServiceVoidResp{}
	var cmd string
	var err *SErr.APIErr
	if force {
		cmd, err = loadCmdScript(s.commonPath, "mv_force")
	} else {
		cmd, err = loadCmdScript(s.commonPath, "mv")
	}
	if err != nil {
		return resp, err
	}
	// sudo mv %s %s
	cmd = fmt.Sprintf(cmd, src, dst)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("LinuxSSHExecutorServiceCommon=[%s], mv, cmd=[%s], output=[%s]", s, cmd, output)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// FileExists 检查文件是否存在
func (s *LinuxSSHExecutorServiceCommon) FileExists(filepath string) (*ExecutorServiceExistsResp, *SErr.APIErr) {
	resp := &ExecutorServiceExistsResp{}
	pathExistsResp, err := s.PathExists(filepath)
	resp.Output = pathExistsResp.Output
	if err != nil {
		return resp, nil
	}
	if !pathExistsResp.Exists {
		resp.Exists = false
		return resp, nil
	}
	cmd, err := loadCmdScript(s.commonPath, "is_file")
	if err != nil {
		return resp, nil
	}
	// sudo [ -f "%s" ] && echo 1 || echo 0
	cmd = fmt.Sprintf(cmd, filepath)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("UbuntuSSHExecutorService=[%s] FileExists, cmd=[%s], output=[%s]", s, cmd, output)
	if err != nil {
		return resp, err
	}
	if strings.Contains(output, "1") {
		resp.Exists = true
		return resp, nil
	} else {
		resp.Exists = false
		return resp, nil
	}
}

// DirExists 检查文件夹是否存在
func (s *LinuxSSHExecutorServiceCommon) DirExists(dirPath string) (*ExecutorServiceExistsResp, *SErr.APIErr) {
	resp := &ExecutorServiceExistsResp{}
	pathExistsResp, err := s.PathExists(dirPath)
	resp.Output = pathExistsResp.Output
	if err != nil {
		return resp, err
	}
	if !pathExistsResp.Exists {
		resp.Exists = false
		return resp, nil
	}
	cmd, err := loadCmdScript(s.commonPath, "is_dir")
	if err != nil {
		return resp, err
	}
	// [ -d "%s" ] && echo 1 || echo 0
	cmd = fmt.Sprintf(cmd, dirPath)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("LinuxSSHExecutorServiceCommon=[%s] DirExists, cmd=[%s] output=[%s]", s, cmd, output)
	if err != nil {
		return resp, err
	}
	if strings.Contains(output, "1") {
		resp.Exists = true
		return resp, nil
	} else {
		resp.Exists = false
		return resp, nil
	}
}

func (s *LinuxSSHExecutorServiceCommon) PathExists(path string) (*ExecutorServiceExistsResp, *SErr.APIErr) {
	resp := &ExecutorServiceExistsResp{}
	cmd, err := loadCmdScript(s.commonPath, "path_exists")
	if err != nil {
		return resp, err
	}
	// ([ -f "%s" ] || [ -d "%s" ]) && echo 1 || echo 0
	cmd = fmt.Sprintf(cmd, path, path)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("LinuxSSHExecutorServiceCommon=[%s] PathExists, cmd=[%s] output=[%s]", s, cmd, output)
	if err != nil {
		return resp, err
	}
	if strings.Contains(output, "1") {
		resp.Exists = true
		return resp, err
	} else {
		resp.Exists = false
		return resp, nil
	}
}


// Mkdir 创建文件夹，不检查文件夹是否存在。
func (s *LinuxSSHExecutorServiceCommon) Mkdir(path string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	resp := &ExecutorServiceVoidResp{}
	cmd, err := loadCmdScript(s.commonPath, "mkdir")
	if err != nil {
		return resp, err
	}
	cmd = fmt.Sprintf(cmd, path)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// MkdirIfNotExists 顾名思义
func (s *LinuxSSHExecutorServiceCommon) MkdirIfNotExists(dirPath string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	resp := &ExecutorServiceVoidResp{}
	pathExistsResp, err := s.PathExists(dirPath)
	resp.Output = pathExistsResp.Output
	if err != nil {
		return resp, err
	}
	if !pathExistsResp.Exists {
		resp, err := s.Mkdir(dirPath)
		if err != nil {
			return resp, err
		}
	}
	return resp, nil
}

// Top 获取CPU，Mem占用，以及Process占用的信息。
func (s *LinuxSSHExecutorServiceCommon) Top() (*ExecutorServiceTopResp, *SErr.APIErr) {
	// top - 08:32:57 up 12 days,  5:24,  5 users,  load average: 1.86, 2.28, 2.49
	// Tasks: 620 total,   1 running, 471 sleeping,   0 stopped,   0 zombie
	// %Cpu(s): 11.1 us,  3.8 sy,  0.9 ni, 83.0 id,  0.1 wa,  0.0 hi,  1.0 si,  0.0 st
	// KiB Mem : 13173032+total, 54938332 free, 45122164 used, 31669828 buff/cache
	// KiB Swap:        0 total,        0 free,        0 used. 86858088 avail Mem
	//
	//  PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
	// 12029 root      20   0  781372 101180   4468 S 100.0  0.1   6101:44 kube-sched
	//  8743 yzc       20   0   43328   4232   3268 R  16.7  0.0   0:00.04 top
	// 29049 root      20   0 5995712 227008  60504 S  11.1  0.2   4261:03 kubelet
	//  4500 root      20   0 1595572 789692  72300 S   5.6  0.6   1723:38 kube-apiserver
	//  4630 onceas    20   0 43.655g 0.032t  32840 S   5.6 26.2 581:32.62 java
	//  5459 onceas    20   0 1261532 412984  39464 S   5.6  0.3 153:25.87 node
	// 15440 root      20   0 3057972  54540  30004 S   5.6  0.0 386:14.40 calico-node
	// 15690 onceas    20   0 1939508 647124  87040 S   5.6  0.5 690:20.78 prometheus
	// 17364 1337      20   0  199440  55632  29568 S   5.6  0.0  38:23.58 envoy
	// 22228 1337      20   0  199444  55848  29520 S   5.6  0.0  38:58.67 envoy
	//     1 root      20   0   80592  11652   6588 S   0.0  0.0  92:23.60 systemd
	resp := &ExecutorServiceTopResp{}
	cmd, err := loadCmdScript(s.commonPath, "top")
	if err != nil {
		return resp, err
	}
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	if err != nil {
		return resp, err
	}

	resp.ProcessInfos = make([]*internal_models.ServerProcessInfo, 0)
	resp.CPUMemUsage = &internal_models.ServerCPUMemUsage{
		UserProcCPUUsage: "",
		MemUsage:         "",
	}
	lines := util.SplitLine(output)
	matchCPU := func(line string) {
		if resp.CPUMemUsage.UserProcCPUUsage != "" {
			return
		}
		reg := regexp.MustCompile(`^.*Cpu\(s\):\s+([0-9.]+) us.*$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 2 {
			return
		}
		log.Printf("LinuxSSHExecutorServiceCommon Top matchCPU=[%+v]", util.Pretty(m))
		f, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			log.Printf("LinuxSSHExecutorServiceCommon Top matchCPU=[%+v], parseFloat failed, err=[%+v]", util.Pretty(m), err)
			return
		}
		resp.CPUMemUsage.UserProcCPUUsage = fmt.Sprintf("%.2f%%", 100*f)
	}
	matchMem := func(line string) {
		if resp.CPUMemUsage.MemUsage != "" {
			return
		}
		reg := regexp.MustCompile(`^.*Mem.*:.*,\s+([0-9]+) free,\s+([0-9]+) used,\s+([0-9]+) buff/cache.*$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 4 {
			return
		}
		log.Printf("LinuxSSHExecutorServiceCommon Top matchMem=[%+v]", util.Pretty(m))
		free, _ := util.ParseInt(m[1])
		used, _ := util.ParseInt(m[2])
		buff, _ := util.ParseInt(m[3])
		if free == 0 || used == 0 || buff == 0 {
			log.Printf("LinuxSSHExecutorServiceCommon Top matchMem=[%+v], parseInt return 0", m)
			return
		}
		resp.CPUMemUsage.MemUsage = fmt.Sprintf("%.2f%%", 100*float64(used) / float64(free + used + buff))
	}
	matchProcLine := func(line string) {
		line = strings.TrimSpace(line)
		splits := util.SplitSpaces(line)
		if len(splits) != 12 {
			if len(resp.ProcessInfos) > 0 {
				// 如果已经开始分析Proc行的数据了，但是却没有匹配成功，则改行数据可能出现匹配问题，打log看看
				log.Printf("出现Proc行后，但是匹配失败的：line=[%s]", line)
			}
			return
		}
		// 228 root      19  -1  174840  69976  58876 S  0.0  3.4   0:19.95 systemd-journal
		pidStr := splits[0]
		pid, err := util.ParseInt(pidStr)
		if err != nil {
			return
		}
		user := splits[1]
		cpuUsage := splits[8]
		cpuUsageF, err := strconv.ParseFloat(cpuUsage, 64)
		if err != nil {
			return
		}
		memUsage := splits[9]
		memUsageF, err := strconv.ParseFloat(memUsage, 64)
		command := splits[11]
		resp.ProcessInfos = append(resp.ProcessInfos, &internal_models.ServerProcessInfo{
			PID:              uint(pid),
			Command:          command,
			OwnerAccountName: user,
			CPUUsage:         fmt.Sprintf("%.2f%%", cpuUsageF),
			MemUsage:         fmt.Sprintf("%.2f%%", memUsageF),
			GPUUsage:         "", // TODO
		})
		// reg := regexp.MustCompile(`^([0-9]+)\s+([a-zA-z_]+)\s+[0-9]+\s+[0-9]+\s+[0-9]+\s+[0-9]+\s+[0-9]+\s+\w+\s+([0-9.]+)\s+([0-9.]+)\s+[0-9:.]+\s+(\w+)$`)
		// m := reg.FindStringSubmatch(line)
		//if len(m) < 6 {
		//	if len(resp.ProcessInfos) > 0 {
		//		// 如果已经开始分析Proc行的数据了，但是却没有匹配成功，则改行数据可能出现匹配问题，打log看看
		//		log.Printf("出现Proc行后，但是匹配失败的：line=[%s]", line)
		//	}
		//	return
		//}

	}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		matchCPU(line)
		matchMem(line)
		matchProcLine(line)
	}

	return resp, nil
}

// getBackupDir 获取备份文件夹路径。目前，就简单备份到/backup目录下，如果不存在则创建。
func (s *LinuxSSHExecutorServiceCommon) getBackupDir(homeDir string) (string, string, *SErr.APIErr) {
	backupDirPath := "/backup"
	mkdirResp, err := s.MkdirIfNotExists(backupDirPath)
	if err != nil {
		return mkdirResp.Output, "", err
	}
	dirElem := path.Base(homeDir)
	targetDirElem := fmt.Sprintf("%s.backup", dirElem)
	return mkdirResp.Output, path.Join(backupDirPath, targetDirElem), nil
}

// BackupAccountHomeDir 将某用户的用户文件夹mv到一份备份。
func (s *LinuxSSHExecutorServiceCommon) BackupAccountHomeDir(accountName string) (*ExecutorServiceBackupAccountResp, *SErr.APIErr) {
	resp := &ExecutorServiceBackupAccountResp{}
	getAccountHomeDirResp, err := s.GetAccountHomeDir(accountName)
	resp.Output = getAccountHomeDirResp.Output
	if err != nil {
		return resp, err
	}
	dirExists, err := s.DirExists(getAccountHomeDirResp.HomeDir)
	resp.Output = dirExists.Output
	if err != nil {
		return resp, err
	}
	if !dirExists.Exists {
		return resp, SErr.BackupDirNotExists.CustomMessageF("备份用户文件夹时，该用户的home目录文件夹不存在！该目录为%s", getAccountHomeDirResp.HomeDir)
	}

	output, targetBackupDirPath, err := s.getBackupDir(getAccountHomeDirResp.HomeDir)
	resp.Output = output
	if err != nil {
		return resp, err
	}

	pathExistsResp, err := s.PathExists(targetBackupDirPath)
	resp.Output = pathExistsResp.Output
	if err != nil {
		return resp, err
	}
	if pathExistsResp.Exists {
		return resp, SErr.BackupTargetDirAlreadyExists.CustomMessageF("备份用户文件夹时，目标的文件夹被占用！该目标路径为：%s", targetBackupDirPath)
	}
	moveResp, err := s.Move(getAccountHomeDirResp.HomeDir, targetBackupDirPath, false)
	resp.Output = moveResp.Output
	if err != nil {
		return resp, err
	}
	resp.TargetDir = targetBackupDirPath
	return resp, nil
}

// RecoverAccountHomeDir 恢复备份的用户目录文件夹。调用该函数时需要保证该用户是存在的，否则无法使用GetAccountHomeDir获取到用户的home目录
func (s *LinuxSSHExecutorServiceCommon) RecoverAccountHomeDir(accountName string, force bool) (*ExecutorServiceRecoverAccountResp, *SErr.APIErr) {
	resp := &ExecutorServiceRecoverAccountResp{}
	getAccountHomeDirResp, err := s.GetAccountHomeDir(accountName)
	resp.Output = getAccountHomeDirResp.Output
	if err != nil {
		return resp, err
	}
	pathExistsResp, err := s.PathExists(getAccountHomeDirResp.HomeDir)
	resp.Output = pathExistsResp.Output
	if err != nil {
		return resp, err
	}
	if pathExistsResp.Exists && !force {
		return resp, SErr.BackupTargetDirAlreadyExists.CustomMessageF("您要恢复到的home文件夹已被占用！请避免覆盖数据！该目标路径为%s", getAccountHomeDirResp.HomeDir)
	}
	output, backupDirPath, err := s.getBackupDir(getAccountHomeDirResp.HomeDir)
	resp.Output = output
	if err != nil {
		return resp, err
	}
	pathExistsResp, err = s.PathExists(backupDirPath)
	resp.Output = pathExistsResp.Output
	if err != nil {
		return resp, err
	}
	resp.Output = pathExistsResp.Output
	if !pathExistsResp.Exists {
		return resp, SErr.BackupDirNotExists.CustomMessageF("恢复用户的目录文件夹时，该备份的文件夹不存在！其路径为：%s", backupDirPath)
	}
	moveResp, err := s.Move(backupDirPath, getAccountHomeDirResp.HomeDir, false)
	resp.Output = moveResp.Output
	if err != nil {
		return resp, err
	}
	resp.HomeDir = getAccountHomeDirResp.HomeDir
	return resp, nil
}

func (s *LinuxSSHExecutorServiceCommon) AddAccount(accountName, pwd string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	switch s.OSType {
	case Ubuntu:
		return s.ubuntuService.AddAccount(accountName, pwd)
	default:
		panic("Unimplemented")
	}
}

func (s *LinuxSSHExecutorServiceCommon) DeleteAccount(accountName string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	switch s.OSType {
	case Ubuntu:
		return s.ubuntuService.DeleteAccount(accountName)
	default:
		panic("Unimplemented")
	}
}

func (s *LinuxSSHExecutorServiceCommon) GetAccountHomeDir(accountName string) (*ExecutorServiceGetAccountHomeDirResp, *SErr.APIErr) {
	switch s.OSType {
	case Ubuntu:
		return s.ubuntuService.GetAccountHomeDir(accountName)
	default:
		panic("Unimplemented")
	}
}

func (s *LinuxSSHExecutorServiceCommon) GetAccountList() (*ExecutorServiceGetAccountListResp, *SErr.APIErr) {
	switch s.OSType {
	case Ubuntu:
		return s.ubuntuService.GetAccountList()
	default:
		panic("Unimplemented")
	}
}

type UbuntuSSHExecutorService struct {
	SSHConn *LinuxSSHConnection
	ubuntuPath string
}

func NewUbuntuSSHExecutorService(SSHConn *LinuxSSHConnection) *UbuntuSSHExecutorService {
	pCmdScrPath := config.GetConfig().CmdsScriptsPath
	ubuntu := path.Join(pCmdScrPath, "ubuntu")
	return &UbuntuSSHExecutorService{
		SSHConn: SSHConn,
		ubuntuPath: ubuntu,
	}
}

func (s *UbuntuSSHExecutorService) String() string {
	return "UbuntuSSHExecutorService"
}

func (s *UbuntuSSHExecutorService) Close() error {
	return s.SSHConn.Close()
}

// LoadSudoersLines 查询/etc/sudoers文件，返回去掉注释的每行内容
func (s *UbuntuSSHExecutorService) LoadSudoersLines() (string, []string, *SErr.APIErr) {
	// 这里给出一个sudoers文件的样例。需要注意的是，要过滤掉注释的行。
	// #
	// # This file MUST be edited with the 'visudo' command as root.
	// #
	// # Please consider adding local content in /etc/sudoers.d/ instead of
	// # directly modifying this file.
	// #
	// # See the man page for details on how to write a sudoers file.
	// #
	// Defaults	env_reset
	// Defaults	mail_badpass
	// Defaults	secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin"
	//
	// # Host alias specification
	//
	// # User alias specification
	//
	// # Cmnd alias specification
	//
	// # User privilege specification
	// root	ALL=(ALL:ALL) ALL
	//
	// # Members of the admin group may gain root privileges
	// %admin ALL=(ALL) ALL
	//
	// # Allow members of group sudo to execute any command
	// %sudo	ALL=(ALL:ALL) ALL
	//
	// # See sudoers(5) for more information on "#include" directives:
	//
	// #includedir /etc/sudoers.d
	// someuser ALL=(ALL:ALL) ALL
	cmd, err := loadCmdScript(s.ubuntuPath, "cat_sudoers")
	if err != nil {
		return "", nil, err
	}
	output, err := s.SSHConn.SendCommands(cmd)
	if err != nil {
		return "", nil, err
	}
	log.Printf("UbuntuSSHExecutorService GetSudoersList cmd=[%s], output=[%s]", cmd, output)
	lines := util.SplitLine(output)
	validLines := make([]string, 0, 4)
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// 注释掉的行，无视掉。
			continue
		}
		validLines = append(validLines, line)
	}
	return "", validLines, nil
}

// Add2Sudoers 为用户添加到sudo权限
func (s *UbuntuSSHExecutorService) Add2Sudoers(accountName string) (string, *SErr.APIErr) {
	output, lines, err := s.LoadSudoersLines()
	if err != nil {
		return output, err
	}
	reg := regexp.MustCompile(`^(.*)?\s+ALL\s*=\s*\(\s*ALL\s*:\s*ALL\s*\)\s*ALL$`)
	for _, line := range lines {
		m := reg.FindStringSubmatch(line)
		if len(m) < 2 {
			continue
		}
		if m[1] == accountName {
			// 在已经有的SudoersFile已经找到了该用户，则放弃本次操作
			log.Printf("UbuntuSSHExecutorService=[%s], Add2Sudoers found account already in sudoers file, line=[%s]", s, line)
			return "", nil
		}
	}

	cmd, err := loadCmdScript(s.ubuntuPath, "add_sudoers")
	cmd = fmt.Sprintf(cmd, accountName)
	if err != nil {
		return "", err
	}
	output, err = s.SSHConn.SendCommands(cmd)
	log.Printf("UbuntuSSHExecutorService=[%s], Add2Sudoers output=[%s]", s, output)
	if err != nil {
		return output, err
	}
	return output, nil
}

// AddAccount 添加一个用户，在该用户有可能是重复的情况下添加。
func (s *UbuntuSSHExecutorService) AddAccount(accountName, pwd string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	resp := &ExecutorServiceVoidResp{}
	if pwd == "" || accountName == "" {
		panic("UbuntuSSHExecutorService AddAccount should have valid param")
	}
	cmd, err := loadCmdScript(s.ubuntuPath, "pwd_digest")
	if err != nil {
		return resp, err
	}
	// 第一步，使用perl生成加密后的密码。
	cmd = fmt.Sprintf(cmd, pwd)
	digestPwd, err := s.SSHConn.SendCommands(cmd)
	log.Printf("UbuntuSSHExecutorService=[%s], AddAccount, pwd_digest output=[%s]", s, digestPwd)
	if err != nil {
		return resp, err
	}
	// 第二步，使用生成的密码添加用户
	cmd, err = loadCmdScript(s.ubuntuPath, "user_add")
	if err != nil {
		return resp, err
	}
	// useradd -m -p "%s" "%s" 需要格式化Pwd以及Name
	cmd = fmt.Sprintf(cmd, digestPwd, accountName)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("UbuntuSSHExecutorService=[%s], AddAccount, user_add cmd output=[%s]", s, output)
	if err != nil {
		return resp, err
	}

	// 第三步，将用户添加sudo权限。
	output, err = s.Add2Sudoers(accountName)
	resp.Output = output
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *UbuntuSSHExecutorService) GetAccountList() (*ExecutorServiceGetAccountListResp, *SErr.APIErr) {
	resp := &ExecutorServiceGetAccountListResp{}
	cmd, err := loadCmdScript(s.ubuntuPath, "get_account_list")
	if err != nil {
		return resp, err
	}
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("UbuntuSSHExecutorService=[%s] GetAccountList, output=[%s]", s, output)
	if err != nil {
		return resp, err
	}
	// output format: account line by line：UserName|UID|GID
	reg := regexp.MustCompile(`^(.+?)\|([0-9]+?)\|([0-9]+?)$`)

	lines := util.SplitLine(output)
	accounts := make([]*internal_models.Account, 0, 4)
	for _, line := range lines {
		log.Printf("UbuntuSSHExecutorService=[%s] GetAccountList, line=[%s]", s, line)
		subs := reg.FindStringSubmatch(line)
		if len(subs) < 4 {
			continue
		}
		account := subs[1]
		uidStr := subs[2]
		gidStr := subs[3]
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			continue
		}
		gid, err := strconv.Atoi(gidStr)
		if err != nil {
			continue
		}
		accounts = append(accounts, &internal_models.Account{
			Host: s.SSHConn.Host,
			Port: s.SSHConn.Port,
			Name: account,
			UID:  uint(uid),
			GID:  uint(gid),
		})
	}
	resp.Accounts = accounts
	return resp, nil
}

// DeleteAccount 删除Linux账户，需注意，删除账户后，使用GetAccountHomeDir会找不到该用户的home目录。所以需要备份的话，需要在删除账户之前做。
func (s *UbuntuSSHExecutorService) DeleteAccount(accountName string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	resp := &ExecutorServiceVoidResp{}
	cmd, err := loadCmdScript(s.ubuntuPath, "user_del")
	if err != nil {
		return resp, err
	}
	// userdel %s
	cmd = fmt.Sprintf(cmd, accountName)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// GetAccountHomeDir 获取账户的home目录
func (s *UbuntuSSHExecutorService) GetAccountHomeDir(accountName string) (*ExecutorServiceGetAccountHomeDirResp, *SErr.APIErr) {
	resp := &ExecutorServiceGetAccountHomeDirResp{}
	cmd, err := loadCmdScript(s.ubuntuPath, "get_user_home_dir")
	if err != nil {
		return resp, err
	}
	// getent passwd "%s" | cut -d: -f6
	cmd = fmt.Sprintf(cmd, accountName)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("UbuntuSSHExecutorService=[%s] GetAccountHomeDir, output=[%s]", s, output)
	if err != nil {
		return resp, err
	}
	homeDir := strings.TrimSpace(output)
	resp.HomeDir = homeDir
	return resp, nil
}

// LinuxSSHConnection SSH的底层连接。
type LinuxSSHConnection struct {
	*ssh.Client
	Password string
	Host string
	Port uint
	Account string
}

func ConnectLinuxSSH(Host string, Port uint, account, password string) (*LinuxSSHConnection, *SErr.APIErr) {
	sshConfig := &ssh.ClientConfig{
		User: account,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", Host, Port), sshConfig)
	if err != nil {
		return nil, SErr.SSHConnectionErr.CustomMessageF("连接ssh失败！错误信息为%s", err.Error())
	}

	return &LinuxSSHConnection{conn, password, Host, Port, account}, nil
}

func (conn *LinuxSSHConnection) String() string {
	return fmt.Sprintf("LinuxSSHConnection=[Host=%s, Port=%d, Account=%s, Password=%s]", conn.Host, conn.Port, conn.Account, conn.Password)
}

func (conn *LinuxSSHConnection) Close() error {
	return conn.Client.Close()
}

// CheckSudoPrivilege 检查该连接的用户是否具有sudo权限。
func (conn *LinuxSSHConnection) CheckSudoPrivilege() (string, bool, *SErr.APIErr) {
	commonCmdPath := path.Join(config.GetConfig().CmdsScriptsPath, "common")
	cmd, err := loadCmdScript(commonCmdPath, "sudo_privilege")
	if err != nil {
		return "", false, err
	}
	output, err := conn.SendCommands(cmd)
	if err != nil {
		return output, false, err
	}
	return output, true, nil
}

// CheckOSInfo 目前只包含操作系统类型，之后可能会针对操作系统版本做改进。
func (conn *LinuxSSHConnection) CheckOSInfo() (string, LinuxOSType, *SErr.APIErr) {
	commonCmdPath := path.Join(config.GetConfig().CmdsScriptsPath, "common")
	cmd, err := loadCmdScript(commonCmdPath, "os_info")
	if err != nil {
		return "", "", err
	}
	output, err := conn.SendCommands(cmd)
	if err != nil {
		return output, "", err
	}
	lower := strings.ToLower(output)
	var osType LinuxOSType
	if strings.Contains(lower, "ubuntu") {
		osType = Ubuntu
	} else if strings.Contains(lower, "centos") {
		osType = CentOS
	} else {
		osType = Unknown
	}
	return output, osType, nil
}

func (conn *LinuxSSHConnection) SendCommandsNoSudo(cmds ...string) (string, *SErr.APIErr) {
	session, err := conn.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	cmd := strings.Join(cmds, "; ")
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), SErr.SSHConnectionErr.CustomMessageF("发送请求后，返回失败信息！失败信息为：%s", err.Error())
	}

	return string(output), nil
}

func (conn *LinuxSSHConnection) SendCommands(cmds ...string) (string, *SErr.APIErr) {
	session, err := conn.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		// ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 400, 400, modes)
	if err != nil {
		return "", SErr.SSHConnectionErr.CustomMessageF("发送ssh命令失败！失败信息为：%s", err.Error())
	}

	in, err := session.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	out, err := session.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	//errOut, err := session.StderrPipe()
	//if err != nil {
	//	log.Fatal(err)
	//}

	//var errOutput []byte
	var output []byte
	wg := &sync.WaitGroup{}
	wg.Add(1)
	//go func(in io.WriteCloser, out io.Reader, output *[]byte) {
	//	var (
	//		r = bufio.NewReader(out)
	//	)
	//	for {
	//		b, err := r.ReadByte()
	//		if err != nil {
	//			break
	//		}
	//		*output = append(*output, b)
	//	}
	//	wg.Done()
	//}(in, errOut, &errOutput)
	go func(in io.WriteCloser, out io.Reader, output *[]byte) {
		var (
			line string
			r    = bufio.NewReader(out)
		)
		for {
			b, err := r.ReadByte()
			if err != nil {
				break
			}

			*output = append(*output, b)

			if b == byte('\n') {
				line = ""
				continue
			}

			line += string(b)

			if strings.HasPrefix(line, "[sudo] password for ") && strings.HasSuffix(line, ": ") {
				_, err = in.Write([]byte(conn.Password + "\n"))
				if err != nil {
					break
				}
			}
		}
		wg.Done()
	}(in, out, &output)

	cmd := strings.Join(cmds, "; ")
	_, err = session.CombinedOutput(cmd)
	if err != nil {
		return string(output), SErr.SSHConnectionErr.CustomMessageF("发送请求后，返回失败信息！失败信息为：%s", err.Error())
	}

	wg.Wait()
	// remove sudo line
	outStr := string(output)
	splits := util.SplitLine(outStr)
	outs := make([]string, 0, len(splits))
	for _, str := range splits {
		if strings.HasPrefix(str, "[sudo] password for ") {
			continue
		}
		outs = append(outs, str)
	}
	//errOutputStr := string(errOutput)
	//log.Printf("SendCommands errOutputStr=[%s]", errOutputStr)
	return strings.Join(outs, "\n"), nil
}