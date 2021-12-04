package internal_models

import (
	"ServerServing/da/mysql/da_models"
	"gorm.io/gorm"
	"time"
)

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
	CPUMemProcessesUsageInfo *ServerCPUMemProcessesUsageInfo `json:"cpu_mem_usage_info"`

	// GPUUsageInfo 当前该Server总的GPU利用率信息。（当前为string，具体待定）
	GPUUsageInfo *ServerGPUUsageInfo `json:"server_gpu_usage_info"`
}

type ServerBasic struct {
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Host             string `json:"host"`
	Port             uint    `json:"port"`
	AdminAccountName string `json:"admin_account_name"`
	AdminAccountPwd  string `json:"admin_account_pwd"`
	OSType da_models.OSType `json:"os_type"`
}

type Account struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Name  string `json:"name"`
	Pwd   string `json:"pwd"`

	Host string `json:"host"`
	Port uint `json:"port"`

	UID uint `json:"uid"`
	GID uint `json:"gid"`

	NotExistsInServer bool `json:"not_exists_in_server"`

	Server ServerBasic `json:"server"`
}

// ServerInfoLoadingFailedInfo 描述一个服务器的某部分内容加载失败的原因，以及当时服务器的原始输出。
type ServerInfoLoadingFailedInfo struct {
	CauseDescription string // 描述具体原因。
}

// ServerInfoCommon 每个Server的信息都要包含的结构。它描述了获取该信息时是否失败，以及对应的服务器原始输出。
// 如果Output为空，则代表该信息并不是独立访问得到的。
// FailedInfo 表示当该部分信息查询失败时的原因。
type ServerInfoCommon struct {
	Output string `json:"output"`
	FailedInfo *ServerInfoLoadingFailedInfo `json:"failed_info"`
}

// ServerHardwareInfo 硬件信息。
type ServerHardwareInfo struct {
	*ServerInfoCommon

	CPUInfos []*ServerCPU `json:"cpu_infos"`
	GPUInfos []*ServerGPU `json:"gpu_infos"`
}

type ServerCPU struct {
	*ServerInfoCommon

	// Summary 描述概要性质的信息，是必选的。
	Summary string `json:"summary"`

	// Type Cores ClockCycle 等描述细节的CPU信息。为空时只关注Summary。
	Type string `json:"type"`
	Cores string `json:"cores"`
	ClockCycle string `json:"clock_cycle"`
}

type ServerGPU struct {
	*ServerInfoCommon

	// Summary 描述概要性质的信息，是必选的。
	Summary string `json:"summary"`

	// Type Cores 等描述细节的GPU硬件信息。为空时只关注Summary。
	Type string `json:"type"`
	Cores string `json:"cores"`
}

type ServerAccountInfos struct {
	*ServerInfoCommon

	Accounts []*Account `json:"accounts"`
}

// ServerRemoteAccessingUsageInfo 描述一个正在从远程访问的用户信息。（正在使用SSH连接的用户）
type ServerRemoteAccessingUsageInfo struct {
	*ServerInfoCommon

	AccessingAccounts []*Account `json:"accessing_accounts"`
}

// ServerCPUMemProcessesUsageInfo 记录当前全部进程的CPU，内存，利用率。
// 实际就是用Top指令获取一个快照。
type ServerCPUMemProcessesUsageInfo struct {
	*ServerInfoCommon

	// CPUMemUsage 服务器总的CPU，内存使用率。
	CPUMemUsage *ServerCPUMemUsage

	// ProcessInfos 全部进程信息。
	ProcessInfos []*ServerProcessInfo `json:"process_infos"`
}

type ServerCPUMemUsage struct {
	// UserProcCPUUsage 记录用户进程的CPU使用率。（总比例）
	UserProcCPUUsage string `json:"user_cpu_usage"`

	// MemUsage 总内存使用（比例：如3600MB/8000MB）
	MemUsage string `json:"mem_usage"`
}

// ServerProcessInfo 描述一个在Server上的进程信息。
type ServerProcessInfo struct {
	*ServerInfoCommon

	// PID 进程号。
	PID uint `json:"pid"`
	// Command 命令
	Command string `json:"command"`
	// OwnerAccountName 该进程由哪个用户启动。
	OwnerAccountName string `json:"owner_account_name"`
	// CPU利用率。
	CPUUsage string `json:"cpu_usage"`
	// 内存利用率
	MemUsage string `json:"mem_usage"`
	// GPU利用率（不一定能查到）
	GPUUsage string `json:"gpu_usage"`
}
// ServerGPUUsageInfo 记录当前GPU使用率。
type ServerGPUUsageInfo struct {
	*ServerInfoCommon

	Summary string `json:"summary"`

	UsageSummaryMap map[*ServerGPU]string `json:"usage_summary_map"`
}

