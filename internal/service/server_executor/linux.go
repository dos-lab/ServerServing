package server_executor

import (
	"ServerServing/config"
	SErr "ServerServing/err"
	"ServerServing/internal/internal_models"
	"ServerServing/util"
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// openLinuxSSHExecutorService
// 获取一个到某服务器的SSH Executor服务实例。建立新的ssh连接。在一次http请求中复用这个服务可以减少连接的创建数量。
// 目前不确定是否可以在多个请求间共享，暂时不太建议这么做（需要测试）。这容易引起并发问题，反正总的并发量不高，目前先随便做一做。
func openLinuxSSHExecutorService(host string, port uint, account, pwd string) (ExecutorService, *SErr.APIErr) {
	SSHConn, err := openLinuxSSHConnection(host, port, account, pwd)
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
	var impl ExecutorService
	switch osType {
	case Ubuntu:
		// golang 的模板方法需要上下两层分别持有引用。
		template := NewLinuxSSHExecutorServiceTemplate(Ubuntu, account, pwd, SSHConn)
		svc := NewUbuntuSSHExecutorService()
		template.implement = svc
		svc.LinuxSSHExecutorServiceTemplate = template
		impl = svc
	case CentOS:
		template := NewLinuxSSHExecutorServiceTemplate(CentOS, account, pwd, SSHConn)
		svc := NewCentOSSSHExecutorService()
		template.implement = svc
		svc.LinuxSSHExecutorServiceTemplate = template
		impl = svc
	default:
		panic("Unsupported LinuxOSType")
	}
	return impl, nil
}

type LinuxSSHExecutorServiceTemplate struct {
	*executorServiceCommon

	Host    string
	Port    uint
	Account string
	Pwd     string

	OSType     LinuxOSType
	SSHConn    *LinuxSSHConnection
	commonPath string
	ubuntuPath string
	centosPath string

	implement ExecutorService
}

func NewLinuxSSHExecutorServiceTemplate(osType LinuxOSType, Account string, Pwd string, SSHConn *LinuxSSHConnection) *LinuxSSHExecutorServiceTemplate {
	csp := config.GetConfig().CmdsScriptsPath
	ubuntu := path.Join(csp, "ubuntu")
	common := path.Join(csp, "linux_common")
	centos := path.Join(csp, "centos")
	return &LinuxSSHExecutorServiceTemplate{
		executorServiceCommon: &executorServiceCommon{},
		Host:                  SSHConn.Host,
		Port:                  SSHConn.Port,
		Account:               Account,
		Pwd:                   Pwd,
		OSType:                osType,
		SSHConn:               SSHConn,
		commonPath:            common,
		ubuntuPath:            ubuntu,
		centosPath:            centos,
	}

}

func (s *LinuxSSHExecutorServiceTemplate) String() string {
	return fmt.Sprintf("LinuxSSHExecutorServiceTemplate=[Host=%s, Port=%d, ServerAccount=%s, Pwd=%s, OSType=%s]", s.Host, s.Port, s.Account, s.Pwd, s.OSType)
}

// Move 移动文件或文件夹
func (s *LinuxSSHExecutorServiceTemplate) Move(src, dst string, force bool) (*ExecutorServiceVoidResp, *SErr.APIErr) {
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
	log.Printf("LinuxSSHExecutorServiceTemplate=[%s], mv, cmd=[%s], output=[%s]", s, cmd, output)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// FileExists 检查文件是否存在
func (s *LinuxSSHExecutorServiceTemplate) FileExists(filepath string) (*ExecutorServiceExistsResp, *SErr.APIErr) {
	resp := &ExecutorServiceExistsResp{}
	pathExistsResp, err := s.implement.PathExists(filepath)
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
func (s *LinuxSSHExecutorServiceTemplate) DirExists(dirPath string) (*ExecutorServiceExistsResp, *SErr.APIErr) {
	resp := &ExecutorServiceExistsResp{}
	pathExistsResp, err := s.implement.PathExists(dirPath)
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
	log.Printf("LinuxSSHExecutorServiceTemplate=[%s] DirExists, cmd=[%s] output=[%s]", s, cmd, output)
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

func (s *LinuxSSHExecutorServiceTemplate) PathExists(path string) (*ExecutorServiceExistsResp, *SErr.APIErr) {
	resp := &ExecutorServiceExistsResp{}
	cmd, err := loadCmdScript(s.commonPath, "path_exists")
	if err != nil {
		return resp, err
	}
	// ([ -f "%s" ] || [ -d "%s" ]) && echo 1 || echo 0
	cmd = fmt.Sprintf(cmd, path, path)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("LinuxSSHExecutorServiceTemplate=[%s] PathExists, cmd=[%s] output=[%s]", s, cmd, output)
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
func (s *LinuxSSHExecutorServiceTemplate) Mkdir(path string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
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
func (s *LinuxSSHExecutorServiceTemplate) MkdirIfNotExists(dirPath string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	resp := &ExecutorServiceVoidResp{}
	pathExistsResp, err := s.implement.PathExists(dirPath)
	resp.Output = pathExistsResp.Output
	if err != nil {
		return resp, err
	}
	if !pathExistsResp.Exists {
		resp, err := s.implement.Mkdir(dirPath)
		if err != nil {
			return resp, err
		}
	}
	return resp, nil
}

// GetCPUMemProcessesUsages 获取CPU，Mem占用，以及Process占用的信息。
func (s *LinuxSSHExecutorServiceTemplate) GetCPUMemProcessesUsages() (*ExecutorServiceCPUMemProcessesUsagesResp, *SErr.APIErr) {
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
	resp := &ExecutorServiceCPUMemProcessesUsagesResp{}
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
		UserProcCPUUsage: nil,
		MemUsage:         nil,
	}
	lines := util.SplitLine(output)
	matchCPU := func(line string) {
		if resp.CPUMemUsage.UserProcCPUUsage != nil {
			return
		}
		reg := regexp.MustCompile(`^.*Cpu\(s\):\s+([0-9.]+) us.*$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 2 {
			return
		}
		log.Printf("LinuxSSHExecutorServiceTemplate GetCPUMemProcessesUsages matchCPU=[%+v]", util.Pretty(m))
		f, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			log.Printf("LinuxSSHExecutorServiceTemplate GetCPUMemProcessesUsages matchCPU=[%+v], parseFloat failed, err=[%+v]", util.Pretty(m), err)
			return
		}
		resp.CPUMemUsage.UserProcCPUUsage = &f
	}
	matchMem := func(line string) {
		if resp.CPUMemUsage.MemUsage != nil {
			return
		}
		reg := regexp.MustCompile(`^.*Mem.*:.*,\s+([0-9]+) free,\s+([0-9]+) used,\s+([0-9]+) buff/cache.*$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 4 {
			return
		}
		log.Printf("LinuxSSHExecutorServiceTemplate GetCPUMemProcessesUsages matchMem=[%+v]", util.Pretty(m))
		free, _ := util.ParseInt(m[1])
		used, _ := util.ParseInt(m[2])
		buff, _ := util.ParseInt(m[3])
		if free == 0 || used == 0 || buff == 0 {
			log.Printf("LinuxSSHExecutorServiceTemplate GetCPUMemProcessesUsages matchMem=[%+v], parseInt return 0", m)
			return
		}
		usage := 100 * float64(used) / float64(free+used+buff)
		total := float64(free + used + buff)
		resp.CPUMemUsage.MemUsage = &usage
		totalStr := strconv.Itoa(int(total/1024.)) + "MB"
		resp.CPUMemUsage.MemTotal = &totalStr
	}
	matchProcLine := func(line string) {
		line = strings.TrimSpace(line)
		splits := util.SplitSpaces(line)
		if len(splits) != 12 {
			if len(resp.ProcessInfos) > 0 {
				// 如果已经开始分析Proc行的数据了，但是却没有匹配成功，则此行数据可能出现匹配问题，打log看看
				log.Printf("出现Proc行后，但是匹配失败的：line=[%s]", line)
			}
			return
		}
		// 228 root      19  -1  174840  69976  58876 S  0.0  3.4   0:19.95 systemd-journal
		pidStr := splits[0]
		pidInt, err := util.ParseInt(pidStr)
		if err != nil {
			return
		}
		pid := uint(pidInt)
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
			PID:              &pid,
			Command:          &command,
			OwnerAccountName: &user,
			CPUUsage:         &cpuUsageF,
			MemUsage:         &memUsageF,
			GPUUsage:         nil, // TODO
		})
	}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		matchCPU(line)
		matchMem(line)
		matchProcLine(line)
	}
	if resp.CPUMemUsage.MemUsage == nil {
		// use cat /proc/meminfo
		log.Printf("LinuxSSHExecutorServiceTemplate GetCPUMemProcessesUsages use /proc/meminfo")
		cmd, err = loadCmdScript(s.commonPath, "meminfo")
		if err != nil {
			return resp, err
		}
		output, err := s.SSHConn.SendCommands(cmd)
		resp.Output += fmt.Sprintf("--- meminfo ---\n%s", output)
		if err != nil {
			log.Printf("LinuxSSHExecutorServiceTemplate GetCPUMemProcessesUsages meminfo failed, err=[%+v]", err)
			return resp, err
		}
		lines := util.SplitLine(output)
		totalReg := regexp.MustCompile(`^MemTotal:\s*([0-9]+)\s+kB$`)
		freeReg := regexp.MustCompile(`^MemFree:\s*([0-9]+)\s+kB$`)
		var totalMatched string
		var freeMatched string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			m := totalReg.FindStringSubmatch(line)
			if len(m) == 2 {
				totalMatched = m[1]
			}
			m = freeReg.FindStringSubmatch(line)
			if len(m) == 2 {
				freeMatched = m[1]
			}
		}
		free, _ := util.ParseInt(freeMatched)
		total, _ := util.ParseInt(totalMatched)
		if free == 0 || total == 0 {
			log.Printf("LinuxSSHExecutorServiceTemplate GetCPUMemProcessesUsages meminfo, parseInt return 0")
			return resp, err
		}
		usage := 100 * float64(total - free) / float64(total)
		resp.CPUMemUsage.MemUsage = &usage
		totalStr := strconv.Itoa(total/1024) + "MB"
		resp.CPUMemUsage.MemTotal = &totalStr
	}

	return resp, nil
}

func (s *LinuxSSHExecutorServiceTemplate) GetCPUHardware() (*ExecutorServiceCPUHardwareResp, *SErr.APIErr) {
	// Architecture:        x86_64
	// CPU op-mode(s):      32-bit, 64-bit
	// Byte Order:          Little Endian
	// CPU(s):              1
	// On-line CPU(s) list: 0
	// Thread(s) per core:  1
	// Core(s) per socket:  1
	// Socket(s):           1
	// NUMA node(s):        1
	// Vendor ID:           GenuineIntel
	// CPU family:          6
	// Model:               79
	// Model name:          Intel(R) Xeon(R) CPU E5-2682 v4 @ 2.50GHz
	// Stepping:            1
	// CPU MHz:             2499.996
	// BogoMIPS:            4999.99
	// Hypervisor vendor:   KVM
	// Virtualization type: full
	// L1d cache:           32K
	// L1i cache:           32K
	// L2 cache:            256K
	// L3 cache:            40960K
	// NUMA node0 CPU(s):   0
	resp := &ExecutorServiceCPUHardwareResp{}
	cmd, err := loadCmdScript(s.commonPath, "lscpu")
	if err != nil {
		return resp, err
	}
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	if err != nil {
		return resp, err
	}
	resp.CPU = &internal_models.ServerCPUs{}
	lines := util.SplitLine(output)
	matchArch := func(line string) {
		if resp.CPU.Architecture != nil {
			return
		}
		line = strings.TrimSpace(line)
		reg := regexp.MustCompile(`^Architecture:\s+([^\s]+)$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 2 {
			return
		}
		log.Printf("LinuxSSHExecutorServiceTemplate GetCPUHardware matchArch, line=[%s], m=[%+v]", line, m)
		resp.CPU.Architecture = &m[1]
	}
	matchCores := func(line string) {
		if resp.CPU.Cores != nil {
			return
		}
		line = strings.TrimSpace(line)
		reg := regexp.MustCompile(`^CPU\(s\):\s+([^\s]+)$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 2 {
			return
		}
		log.Printf("LinuxSSHExecutorServiceTemplate GetCPUHardware matchCores, line=[%s], m=[%+v]", line, m)
		cores, err := util.ParseInt(m[1])
		if err != nil {
			return
		}
		resp.CPU.Cores = &cores
	}
	matchThreadsPerCore := func(line string) {
		if resp.CPU.ThreadsPerCore != nil {
			return
		}
		line = strings.TrimSpace(line)
		reg := regexp.MustCompile(`^Thread\(s\) per core:\s+([^\s]+)$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 2 {
			return
		}
		log.Printf("LinuxSSHExecutorServiceTemplate GetCPUHardware matchThreadsPerCore, line=[%s], m=[%+v]", line, m)
		threadsPerCore, err := util.ParseInt(m[1])
		if err != nil {
			return
		}
		resp.CPU.ThreadsPerCore = &threadsPerCore
	}
	matchModelName := func(line string) {
		if resp.CPU.ModelName != nil {
			return
		}
		line = strings.TrimSpace(line)
		reg := regexp.MustCompile(`^Model name:\s+(.*)$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 2 {
			return
		}
		log.Printf("LinuxSSHExecutorServiceTemplate GetCPUHardware matchModelName, line=[%s], m=[%+v]", line, m)
		resp.CPU.ModelName = &m[1]
	}
	matchers := []func(line string){
		matchArch,
		matchCores,
		matchThreadsPerCore,
		matchModelName,
	}
	for _, line := range lines {
		for _, matcher := range matchers {
			matcher(line)
		}
	}
	return resp, nil
}

func (s *LinuxSSHExecutorServiceTemplate) GetGPUHardware() (*ExecutorServiceGPUHardwareResp, *SErr.APIErr) {
	// 17:00.0 VGA compatible controller: NVIDIA Corporation GV102 (rev a1)
	// b3:00.0 VGA compatible controller: NVIDIA Corporation GV102 (rev a1)
	resp := &ExecutorServiceGPUHardwareResp{}
	cmd, err := loadCmdScript(s.commonPath, "lsgpu")
	if err != nil {
		return resp, err
	}
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	if err != nil {
		return resp, err
	}
	resp.GPUs = make([]*internal_models.ServerGPU, 0)
	lines := util.SplitLine(output)
	matchVGA := func(line string) {
		line = strings.TrimSpace(line)
		reg := regexp.MustCompile(`^.*VGA.*controller: (.*)$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 2 {
			return
		}
		log.Printf("LinuxSSHExecutorServiceTemplate GetGPUHardware matchVGA, line=[%s], m=[%+v]", line, m)
		resp.GPUs = append(resp.GPUs, &internal_models.ServerGPU{
			Product: &m[1],
		})
	}
	for _, line := range lines {
		matchVGA(line)
	}

	return resp, nil
}

func (s *LinuxSSHExecutorServiceTemplate) GetMemoryHardware() (*ExecutorServiceMemoryHardwareResp, *SErr.APIErr) {
	resp := &ExecutorServiceMemoryHardwareResp{}
	cmd, err := loadCmdScript(s.commonPath, "meminfo")
	if err != nil {
		return resp, err
	}
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	if err != nil {
		return resp, err
	}
	resp.MemoryStats = &internal_models.ServerMemory{}
	matchTotalMemory := func(line string) {
		if resp.MemoryStats.TotalMemory != nil {
			return
		}
		line = strings.TrimSpace(line)
		reg := regexp.MustCompile(`^MemTotal:\s+(.*)$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 2 {
			return
		}
		resp.MemoryStats.TotalMemory = &m[1]
	}
	lines := util.SplitLine(output)
	for _, line := range lines {
		matchTotalMemory(line)
	}
	return resp, nil
}

// GetBackupDir 获取备份文件夹路径。目前，就简单备份到/backup目录下，如果不存在则创建。
func (s *LinuxSSHExecutorServiceTemplate) GetBackupDir(accountName string) (*ExecutorServiceGetBackupDirResp, *SErr.APIErr) {
	resp := &ExecutorServiceGetBackupDirResp{}
	backupDirPath := "/backup"
	mkdirResp, err := s.implement.MkdirIfNotExists(backupDirPath)
	resp.Output = mkdirResp.Output
	if err != nil {
		return resp, err
	}
	targetDirElem := fmt.Sprintf("%s.backup", accountName)
	targetDir := path.Join(backupDirPath, targetDirElem)
	dirExists, err := s.implement.DirExists(targetDir)
	if err != nil {
		return resp, err
	}
	pathExists, err := s.implement.PathExists(targetDir)
	if err != nil {
		return resp, err
	}
	resp.PathExists = pathExists.Exists
	resp.DirExists = dirExists.Exists
	resp.BackupDir = targetDir
	return resp, nil
}

// BackupAccountHomeDir 将某用户的用户文件夹mv到一份备份。是模板方法
func (s *LinuxSSHExecutorServiceTemplate) BackupAccountHomeDir(accountName string) (*ExecutorServiceBackupAccountResp, *SErr.APIErr) {
	resp := &ExecutorServiceBackupAccountResp{}
	getAccountHomeDirResp, err := s.implement.GetAccountHomeDir(accountName)
	resp.Output = getAccountHomeDirResp.Output
	if err != nil {
		return resp, err
	}
	dirExists, err := s.implement.DirExists(getAccountHomeDirResp.HomeDir)
	resp.Output = dirExists.Output
	if err != nil {
		return resp, err
	}
	if !dirExists.Exists {
		return resp, SErr.BackupDirNotExists.CustomMessageF("备份用户文件夹时，该用户的home目录文件夹不存在！该目录为%s", getAccountHomeDirResp.HomeDir)
	}

	getBackupDirResp, err := s.implement.GetBackupDir(accountName)
	resp.Output = getBackupDirResp.Output
	if err != nil {
		return resp, err
	}

	if getBackupDirResp.PathExists {
		return resp, SErr.BackupTargetDirAlreadyExists.CustomMessageF("备份用户文件夹时，目标的文件夹被占用！该目标路径为：%s", getBackupDirResp.BackupDir)
	}
	moveResp, err := s.implement.Move(getAccountHomeDirResp.HomeDir, getBackupDirResp.BackupDir, false)
	resp.Output = moveResp.Output
	if err != nil {
		return resp, err
	}
	resp.TargetDir = getBackupDirResp.BackupDir
	return resp, nil
}

// RecoverAccountHomeDir 恢复备份的用户目录文件夹。调用该函数时需要保证该用户是存在的，否则无法使用GetAccountHomeDir获取到用户的home目录
func (s *LinuxSSHExecutorServiceTemplate) RecoverAccountHomeDir(accountName string, force bool) (*ExecutorServiceRecoverAccountResp, *SErr.APIErr) {
	resp := &ExecutorServiceRecoverAccountResp{}
	getAccountHomeDirResp, err := s.implement.GetAccountHomeDir(accountName)
	resp.Output = getAccountHomeDirResp.Output
	if err != nil {
		return resp, err
	}
	pathExistsResp, err := s.implement.PathExists(getAccountHomeDirResp.HomeDir)
	resp.Output = pathExistsResp.Output
	if err != nil {
		return resp, err
	}
	if pathExistsResp.Exists && !force {
		return resp, SErr.BackupTargetDirAlreadyExists.CustomMessageF("您要恢复到的home文件夹已被占用！请避免覆盖数据！该目标路径为%s", getAccountHomeDirResp.HomeDir)
	}
	getBackupDirResp, err := s.implement.GetBackupDir(accountName)
	resp.Output = getBackupDirResp.Output
	if err != nil {
		return resp, err
	}
	if !getBackupDirResp.DirExists {
		return resp, SErr.BackupDirNotExists.CustomMessageF("恢复用户的目录文件夹时，该备份的文件夹不存在！其路径为：%s", getBackupDirResp.BackupDir)
	}
	moveResp, err := s.implement.Move(getBackupDirResp.BackupDir, getAccountHomeDirResp.HomeDir, false)
	resp.Output = moveResp.Output
	if err != nil {
		return resp, err
	}
	resp.HomeDir = getAccountHomeDirResp.HomeDir
	return resp, nil
}

func (s *LinuxSSHExecutorServiceTemplate) GetRemoteAccessInfos() (*ExecutorServiceRemoteAccessResp, *SErr.APIErr) {
	resp := &ExecutorServiceRemoteAccessResp{}

	//_, _ = s.SSHConn.SendCommands("export PROCPS_USERLEN=20")
	cmd, err := loadCmdScript(s.commonPath, "w")
	if err != nil {
		return resp, err
	}
	log.Printf("LinuxSSHExecutorServiceTemplate GetRemoteAccessInfos cmd=[%s]", cmd)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("LinuxSSHExecutorServiceTemplate GetRemoteAccessInfos w executed, output=[%s], err=[%v]", output, err)
	if err != nil {
		return resp, err
	}
	matchRemoteAccessing := func(line string) {
		// root     pts/0    114.254.1.92      6.00s sudo w -s -h
		line = strings.TrimSpace(line)
		reg := regexp.MustCompile(`^(\w+)\s+[^\s]+\s+[^\s]+\s+[^\s]+\s+(.*)$`)
		m := reg.FindStringSubmatch(line)
		if len(m) < 3 {
			return
		}
		accountName := m[1]
		what := m[2]
		resp.RemoteAccessingAccountInfos = append(resp.RemoteAccessingAccountInfos, &internal_models.ServerRemoteAccessingAccount{
			AccountName: accountName,
			What:        what,
		})
	}
	resp.RemoteAccessingAccountInfos = make([]*internal_models.ServerRemoteAccessingAccount, 0)
	lines := util.SplitLine(output)
	for _, line := range lines {
		matchRemoteAccessing(line)
	}
	return resp, nil
}

func (s *LinuxSSHExecutorServiceTemplate) GetGPUUsages() (*ExecutorServiceVoidResp, *SErr.APIErr) {
	resp := &ExecutorServiceVoidResp{}
	//gpuHardwareResp, err := s.implement.GetGPUHardware()
	//if err != nil {
	//	return resp, err
	//}
	//// 目前只写CUDA的。别的GPU写了也没用啊！
	//isNvidia := false
	//for _, gpu := range gpuHardwareResp.GPUs {
	//	if gpu.IsNvidia() {
	//		isNvidia = true
	//	}
	//}
	//if isNvidia {
		cmd, err := loadCmdScript(s.commonPath, "nvidia_gpu_name")
		if err != nil {
			return resp, err
		}
		output, _ := s.SSHConn.SendCommands(cmd)
		gpuNames := output
		cmd, err = loadCmdScript(s.commonPath, "nvidia_gpu_usage")
		if err != nil {
			return resp, err
		}
		output, err = s.SSHConn.SendCommands(cmd)
		resp.Output = fmt.Sprintf("%s\n%s", gpuNames, output)
		if err != nil {
			return resp, err
		}
		return resp, nil
	//}
	//log.Printf("LinuxSSHExecutorServiceTemplate GetGPUUsages no cuda gpu, skip.")
	//return resp, nil
}

// LoadSudoersLines 查询/etc/sudoers文件，返回去掉注释的每行内容
func (s *LinuxSSHExecutorServiceTemplate) LoadSudoersLines() (string, []string, *SErr.APIErr) {
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
	cmd, err := loadCmdScript(s.commonPath, "cat_sudoers")
	if err != nil {
		return "", nil, err
	}
	output, err := s.SSHConn.SendCommands(cmd)
	if err != nil {
		return "", nil, err
	}
	log.Printf("LinuxSSHExecutorServiceTemplate GetSudoersList cmd=[%s], output=[%s]", cmd, output)
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
func (s *LinuxSSHExecutorServiceTemplate) Add2Sudoers(accountName string) (string, *SErr.APIErr) {
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
			log.Printf("LinuxSSHExecutorServiceTemplate=[%s], Add2Sudoers found account already in sudoers file, line=[%s]", s, line)
			return "", nil
		}
	}

	cmd, err := loadCmdScript(s.commonPath, "add_sudoers")
	cmd = fmt.Sprintf(cmd, accountName)
	if err != nil {
		return "", err
	}
	output, err = s.SSHConn.SendCommands(cmd)
	log.Printf("LinuxSSHExecutorServiceTemplate=[%s], Add2Sudoers output=[%s]", s, output)
	if err != nil {
		return output, err
	}
	return output, nil
}

// AddAccount 添加一个用户，在该用户有可能是重复的情况下添加。
func (s *LinuxSSHExecutorServiceTemplate) AddAccount(accountName, pwd string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	resp := &ExecutorServiceVoidResp{}
	if pwd == "" || accountName == "" {
		panic("LinuxSSHExecutorServiceTemplate AddAccount should have valid param")
	}
	//cmd, err := loadCmdScript(s.commonPath, "openssl_pwd_digest")
	//if err != nil {
	//	return resp, err
	//}
	// 第一步，使用生成加密后的密码。
	//cmd = fmt.Sprintf(cmd, pwd)
	//digestPwd := s.encrypt(pwd)
	//// digestPwd, err := s.SSHConn.SendCommands(cmd)
	//log.Printf("LinuxSSHExecutorServiceTemplate=[%s], AddAccount, pwd_digest output=[%s]", s, digestPwd)
	//if err != nil {
	//	return resp, err
	//}
	// 第二步，使用生成的密码添加用户
	//cmd, err := loadCmdScript(s.commonPath, "user_add")
	//if err != nil {
	//	return resp, err
	//}
	//// useradd -m -p "%s" "%s" 需要格式化Pwd以及Name
	//cmd = fmt.Sprintf(cmd, digestPwd, accountName)

	// 使用 user_add_with_openssl_pwd
	// sudo useradd -s /bin/bash -m -p $(openssl passwd -crypt "%s") "%s"
	cmd, err := loadCmdScript(s.commonPath, "user_add_with_openssl_pwd")
	cmd = fmt.Sprintf(cmd, pwd, accountName)
	log.Printf("LinuxSSHExecutorServiceTemplate=[%s], AddAccount cmd=[%s]", s, cmd)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("LinuxSSHExecutorServiceTemplate=[%s], AddAccount, user_add cmd output=[%s]", s, output)
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

func (s *LinuxSSHExecutorServiceTemplate) GetAccountList() (*ExecutorServiceGetAccountListResp, *SErr.APIErr) {
	resp := &ExecutorServiceGetAccountListResp{}
	cmd, err := loadCmdScript(s.commonPath, "get_account_list")
	if err != nil {
		return resp, err
	}
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("LinuxSSHExecutorServiceTemplate=[%s] GetAccountList, output=[%s]", s, output)
	if err != nil {
		return resp, err
	}
	// output format: account line by line：UserName|UID|GID
	reg := regexp.MustCompile(`^(.+?)\|([0-9]+?)\|([0-9]+?)$`)

	lines := util.SplitLine(output)
	accounts := make([]*internal_models.ServerAccount, 0, 4)
	for _, line := range lines {
		log.Printf("LinuxSSHExecutorServiceTemplate=[%s] GetAccountList, line=[%s]", s, line)
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
		accounts = append(accounts, &internal_models.ServerAccount{
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
func (s *LinuxSSHExecutorServiceTemplate) DeleteAccount(accountName string) (*ExecutorServiceVoidResp, *SErr.APIErr) {
	resp := &ExecutorServiceVoidResp{}
	cmd, err := loadCmdScript(s.commonPath, "user_del")
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
func (s *LinuxSSHExecutorServiceTemplate) GetAccountHomeDir(accountName string) (*ExecutorServiceGetAccountHomeDirResp, *SErr.APIErr) {
	resp := &ExecutorServiceGetAccountHomeDirResp{}
	cmd, err := loadCmdScript(s.commonPath, "get_user_home_dir")
	if err != nil {
		return resp, err
	}
	// getent passwd "%s" | cut -d: -f6
	cmd = fmt.Sprintf(cmd, accountName)
	output, err := s.SSHConn.SendCommands(cmd)
	resp.Output = output
	log.Printf("LinuxSSHExecutorServiceTemplate=[%s] GetAccountHomeDir, output=[%s]", s, output)
	if err != nil {
		return resp, err
	}
	homeDir := strings.TrimSpace(output)
	resp.HomeDir = homeDir
	return resp, nil
}

func (s *LinuxSSHExecutorServiceTemplate) Close() error {
	return s.SSHConn.Close()
}

type UbuntuSSHExecutorService struct {
	*LinuxSSHExecutorServiceTemplate
	ubuntuPath string
}

func NewUbuntuSSHExecutorService() *UbuntuSSHExecutorService {
	pCmdScrPath := config.GetConfig().CmdsScriptsPath
	ubuntu := path.Join(pCmdScrPath, "ubuntu")
	return &UbuntuSSHExecutorService{
		ubuntuPath: ubuntu,
	}
}

type CentOSSSHExecutorService struct {
	*LinuxSSHExecutorServiceTemplate
	centOSPath string
}

func NewCentOSSSHExecutorService() *CentOSSSHExecutorService {
	pCmdScrPath := config.GetConfig().CmdsScriptsPath
	centos := path.Join(pCmdScrPath, "centos")
	return &CentOSSSHExecutorService{
		centOSPath: centos,
	}
}

// LinuxSSHConnection SSH的底层连接。
type LinuxSSHConnection struct {
	*ssh.Client
	Password string
	Host     string
	Port     uint
	Account  string
}

func openLinuxSSHConnection(Host string, Port uint, account, password string) (*LinuxSSHConnection, *SErr.APIErr) {
	sshConfig := &ssh.ClientConfig{
		Timeout: 10 * time.Second,
		User:    account,
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
	return fmt.Sprintf("LinuxSSHConnection=[Host=%s, Port=%d, ServerAccount=%s, Password=%s]", conn.Host, conn.Port, conn.Account, conn.Password)
}

func (conn *LinuxSSHConnection) Close() error {
	return conn.Client.Close()
}

// CheckSudoPrivilege 检查该连接的用户是否具有sudo权限。
func (conn *LinuxSSHConnection) CheckSudoPrivilege() (string, bool, *SErr.APIErr) {
	commonCmdPath := path.Join(config.GetConfig().CmdsScriptsPath, "linux_common")
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
	commonCmdPath := path.Join(config.GetConfig().CmdsScriptsPath, "linux_common")
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

func (conn *LinuxSSHConnection) SendCommandsNoSudo(envs []string, cmds ...string) (string, *SErr.APIErr) {
	session, err := conn.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = session.Close()
	}()

	cmd := strings.Join(cmds, "; ")
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), SErr.SSHConnectionErr.CustomMessageF("发送请求后，返回失败信息！失败信息为：%s", err.Error())
	}

	return string(output), nil
}

func (conn *LinuxSSHConnection) withSession(envs map[string]string, f func(session *ssh.Session)) *SErr.APIErr {
	session, err := conn.NewSession()
	if err != nil {
		log.Printf("LinuxSSHConnection ssh new session failed, err=[%v]", err)
		return SErr.SSHConnectionErr.CustomMessageF("SSH New Session Failed, err=[%v]", err)
	}
	if envs != nil {
		for key, value := range envs {
			envErr := session.Setenv(key, value)
			if envErr != nil {
				log.Printf("LinuxSSHConnection withSession Setenv failed, key=[%s], value=[%s], err=[%v]", key, value, envErr)
			}
		}
	}
	f(session)
	defer func() {
		_ = session.Close()
	}()
	return nil
}

func (conn *LinuxSSHConnection) SendCommands(cmds ...string) (string, *SErr.APIErr) {
	var output string
	var err *SErr.APIErr
	sessErr := conn.withSession(nil, func(session *ssh.Session) {
		output, err = conn.sendCommandsWithSession(session, cmds...)
	})
	if sessErr != nil {
		return "", SErr.SSHConnectionErr.CustomMessageF("SSH New Session Failed")
	}
	return output, err
}

func (conn *LinuxSSHConnection) SendCommandsWithEnv(envs map[string]string, cmds ...string) (string, *SErr.APIErr) {
	var output string
	var err *SErr.APIErr
	sessErr := conn.withSession(envs, func(session *ssh.Session) {
		output, err = conn.sendCommandsWithSession(session, cmds...)
	})
	if sessErr != nil {
		return "", SErr.SSHConnectionErr.CustomMessageF("SSH New Session Failed")
	}
	return output, err
}

func (conn *LinuxSSHConnection) sendCommandsWithSession(session *ssh.Session, cmds ...string) (string, *SErr.APIErr) {
	modes := ssh.TerminalModes{
		// ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err := session.RequestPty("xterm", 1600, 1600, modes)
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

	var output []byte
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(in io.WriteCloser, out io.Reader, output *[]byte) {
		defer wg.Done()
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
	}(in, out, &output)
	cmd := strings.Join(cmds, "; ")
	// bs, err := session.CombinedOutput(cmd)
	err = session.Run(cmd)
	if err != nil {
		return string(output), SErr.SSHConnectionErr.CustomMessageF("发送请求后，返回失败信息！失败信息为：%s，服务器输出为：%s", err.Error(), string(output))
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
	return strings.Join(outs, "\n"), nil
}
