package internal_models

import (
	"ServerServing/da/mysql/da_models"
	_ "database/sql"
	"gorm.io/gorm"
	"strings"
	"time"
)

type ServerCreateRequest struct {
	Name             string           `form:"name" json:"name"`
	Description      string           `form:"description" json:"description"`
	Host             string           `form:"host" json:"host"`
	Port             uint             `form:"port" json:"port"`
	OSType           da_models.OSType `form:"os_type" json:"os_type"`
	AdminAccountName string           `form:"admin_account_name" json:"admin_account_name"`
	AdminAccountPwd  string           `form:"admin_account_pwd" json:"admin_account_pwd"`
}

type ServerCreateResponse struct {
}

type ServerUpdateRequest struct {
	Name             string `form:"name" json:"name"`
	Description      string `form:"description" json:"description"`
	AdminAccountName string `form:"admin_account_name" json:"admin_account_name"`
	AdminAccountPwd  string `form:"admin_account_pwd" json:"admin_account_pwd"`
}

type ServerUpdateResponse struct {
}

type ServerDeleteRequest struct {
	Host string `form:"host" json:"host"`
	Port uint   `form:"port" json:"port"`
}

type ServerDeleteResponse struct {
}

type ServerInfoRequest struct {
	LoadServerDetailArg
}

type ServerInfoResponse struct {
	*ServerInfo
}

type ServerInfosRequest struct {
	From    uint    `form:"from" json:"from"`
	Size    uint    `form:"size" json:"size"`
	Keyword *string `form:"keyword" json:"keyword"`
	LoadServerDetailArg
}

type ServerInfosResponse struct {
	Infos      []*ServerInfo `json:"infos"`
	TotalCount uint          `json:"total_count"`
}

type LoadServerDetailArg struct {
	// WithHardwareInfo 指定是否加载硬件的元信息
	WithHardwareInfo bool `form:"with_hardware_info" json:"with_hardware_info"`
	// WithAccounts 加载账户信息的参数，为nil则不加载
	WithAccounts bool `form:"with_accounts" json:"with_accounts"`
	// WithAccountsIgnoreDBAccounts 指定是否无视数据库内的账户信息
	WithAccountsIgnoreDBAccounts bool `json:"-" form:"-"`
	// WithRemoteAccessUsages 指定是否加载正在远程登录这台服务器的用户信息。
	WithRemoteAccessUsages bool `form:"with_remote_access_usages" json:"with_remote_access_usages"`
	// WithGPUUsages 指定是否加载GPU的使用信息。
	WithGPUUsages bool `form:"with_gpu_usages" json:"with_gpu_usages"`
	// WithCPUMemProcessesUsage 指定是否加载CPU，内存，进程的使用信息。
	WithCPUMemProcessesUsage bool `form:"with_cmp_usages" json:"with_cmp_usages"`
	// WithBackupDirInfo 指定是否加载用户备份文件夹的信息。
	WithBackupDirInfo bool `form:"with_backup_dir_info" json:"with_backup_dir_info"`
}

// ServerInfo 包含了查询一个Server的详细信息结构体。包含可能的一切数据，从中选取子集展示。
type ServerInfo struct {
	// Basic 基本的Server目录信息
	Basic *ServerBasic `json:"basic"`

	// AccessFailedInfo 指定了当该服务器连接失败时的信息。如果该字段不为空，那么其他字段才有意义。
	AccessFailedInfo *ServerInfoLoadingFailedInfo `json:"access_failed_info"`

	// AccountInfos 记录服务器账户信息。
	AccountInfos *ServerAccountInfos `json:"account_infos"`

	// ServerHardwareInfo 硬件元信息
	HardwareInfo *ServerHardwareInfo `json:"hardware_info"`

	// CPUMemProcessesUsageInfo CPU，内存，进程的使用资源信息。（Top指令）
	CPUMemProcessesUsageInfo *ServerCPUMemProcessesUsageInfo `json:"cpu_mem_processes_usage_info"`

	// RemoteAccessingUsageInfo 正在从远端访问的用户的使用信息
	RemoteAccessingUsageInfo *ServerRemoteAccessingUsagesInfo `json:"remote_accessing_usage_info"`

	// GPUUsageInfo 当前该Server总的GPU利用率信息。（当前为string，具体待定）
	GPUUsageInfo *ServerGPUUsageInfo `json:"server_gpu_usage_info"`
}

type ServerBasic struct {
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Name             string           `json:"name"`
	Description      string           `json:"description"`
	Host             string           `json:"host"`
	Port             uint             `json:"port"`
	AdminAccountName string           `json:"admin_account_name"`
	AdminAccountPwd  string           `json:"admin_account_pwd"`
	OSType           da_models.OSType `json:"os_type"`
}

type ServerAccount struct {
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Name string `json:"name"`
	Pwd  string `json:"pwd"`

	Host string `json:"host"`
	Port uint   `json:"port"`

	UID uint `json:"uid"`
	GID uint `json:"gid"`

	NotExistsInServer bool `json:"not_exists_in_server"`

	BackupDirInfo *ServerAccountBackupDirInfo `json:"backup_dir_info"`

	Server ServerBasic `json:"-"`
}

type ServerAccountBackupDirInfo struct {
	*ServerInfoCommon

	BackupDir  string `json:"backup_dir"`
	PathExists bool   `json:"path_exists"`
	DirExists  bool   `json:"dir_exists"`
}

// ServerInfoLoadingFailedInfo 描述一个服务器的某部分内容加载失败的原因，以及当时服务器的原始输出。
type ServerInfoLoadingFailedInfo struct {
	CauseDescription string `json:"cause_description"` // 描述具体原因。
}

// ServerInfoCommon 每个Server的信息都要包含的结构。它描述了获取该信息时是否失败，以及对应的服务器原始输出。
// 如果Output为空，则代表该信息并不是独立访问得到的。
// FailedInfo 表示当该部分信息查询失败时的原因。
type ServerInfoCommon struct {
	Output     string                       `json:"output"`
	FailedInfo *ServerInfoLoadingFailedInfo `json:"failed_info"`
}

// ServerHardwareInfo 硬件信息。
type ServerHardwareInfo struct {
	*ServerInfoCommon

	CPUHardwareInfo  *ServerCPUHardwareInfo  `json:"cpu_hardware_info"`
	GPUHardwareInfos *ServerGPUHardwareInfos `json:"gpu_hardware_infos"`
}

type ServerCPUHardwareInfo struct {
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
	// Flags:               fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ss ht syscall nx pdpe1gb rdtscp lm constant_tsc rep_good nopl nonstop_tsc cpuid tsc_known_freq pni pclmulqdq ssse3 fma cx16 pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand hypervisor lahf_lm abm 3dnowprefetch invpcid_single pti ibrs ibpb stibp fsgsbase tsc_adjust bmi1 hle avx2 smep bmi2 erms invpcid rtm rdseed adx smap xsaveopt arat
	*ServerInfoCommon

	Info *ServerCPUs `json:"info"`
}

type ServerCPUs struct {
	// Architecture 架构
	Architecture *string `json:"architecture"`
	// ModelName 如：Intel(R) Xeon(R) CPU E5-2682 v4 @ 2.50GHz
	ModelName *string `json:"model_name"`
	// Cores CPU核数
	Cores *int `json:"cores"`
	// ThreadsPerCore 每个核心可以跑几个线程
	ThreadsPerCore *int `json:"threads_per_core"`
}

type ServerGPUHardwareInfos struct {
	*ServerInfoCommon

	Infos []*ServerGPU `json:"infos"`
}

type ServerGPU struct {
	// Product 产品名。
	Product *string `json:"product"`
}

type ServerMemoryHardwareInfo struct {
	*ServerInfoCommon

	MemoryStats *ServerMemory `json:"memory_stats"`
}

type ServerMemory struct {
	TotalMemory *string `json:"total_memory"`
}

func (g ServerGPU) IsNvidia() bool {
	return strings.Contains(strings.ToLower(*g.Product), "nvidia")
}

type ServerAccountInfos struct {
	*ServerInfoCommon

	Accounts []*ServerAccount `json:"accounts"`
}

// ServerRemoteAccessingUsagesInfo 描述一个正在从远程访问的用户信息。（正在使用SSH连接的用户）
type ServerRemoteAccessingUsagesInfo struct {
	*ServerInfoCommon

	Infos []*ServerRemoteAccessingAccount `json:"infos"`
}

// ServerRemoteAccessingAccount 表示一个正在远端访问的用户的信息。
type ServerRemoteAccessingAccount struct {
	AccountName string `json:"account_name"`
	// What 表示该远程访问的用户正在执行的命令。
	What string `json:"what"`
}

// ServerCPUMemProcessesUsageInfo 记录当前全部进程的CPU，内存，利用率。
// 实际就是用Top指令获取一个快照。
type ServerCPUMemProcessesUsageInfo struct {
	*ServerInfoCommon

	// CPUMemUsage 服务器总的CPU，内存使用率。
	CPUMemUsage *ServerCPUMemUsage `json:"cpu_mem_usage"`

	// ProcessInfos 全部进程信息。
	ProcessInfos []*ServerProcessInfo `json:"process_infos"`
}

type ServerCPUMemUsage struct {
	// UserProcCPUUsage 记录用户进程的CPU使用率。（总比例）
	UserProcCPUUsage *float64 `json:"user_cpu_usage"`

	// MemUsage 总内存使用（比例：如3600MB/8000MB）
	MemUsage *float64 `json:"mem_usage"`

	// MemTotal 内存总量，使用字符串固定死
	MemTotal string `json:"mem_total"`
}

// ServerProcessInfo 描述一个在Server上的进程信息。
type ServerProcessInfo struct {
	*ServerInfoCommon

	// PID 进程号。
	PID *uint `json:"pid"`
	// Command 命令
	Command *string `json:"command"`
	// OwnerAccountName 该进程由哪个用户启动。
	OwnerAccountName *string `json:"owner_account_name"`
	// CPU利用率。
	CPUUsage *float64 `json:"cpu_usage"`
	// 内存利用率
	MemUsage *float64 `json:"mem_usage"`
	// GPU利用率（不一定能查到）
	GPUUsage *string `json:"gpu_usage"`
}

// ServerGPUUsageInfo 记录当前GPU使用情况。
// GPU的使用率查询比较复杂，直接展示原输出。
type ServerGPUUsageInfo struct {
	*ServerInfoCommon
}

type ServerConnectionTestRequest struct {
	AccountName string           `form:"account_name" json:"account_name"`
	AccountPwd  string           `form:"account_pwd" json:"account_pwd"`
	OSType      da_models.OSType `form:"os_type" json:"os_type"`
}

type ServerConnectionTestResponse struct {
	Connected bool   `json:"connected"`
	Cause     string `json:"cause"`
}
