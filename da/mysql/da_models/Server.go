package da_models

import (
	"gorm.io/gorm"
	"time"
)

type OSType string

const (
	OSTypeLinux         OSType = "os_type_linux"
	OSTypeWindowsServer OSType = "os_type_windows_server"
)

type Server struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name             string `gorm:"unique;index:idx_servers_name;size:20"`
	Description      string `gorm:"size:140"`
	Host             string `gorm:"primaryKey;index:idx_servers_host_port,priority:1;size:20"`
	Port             uint   `gorm:"primaryKey;index:idx_servers_host_port,priority:2"`
	AdminAccountName string `gorm:"index;not null;size:50"`
	AdminAccountPwd  string `gorm:"not null;size:50"`
	OSType           OSType `json:"os_type" gorm:"not null" sql:"type:ENUM('os_type_linux', 'os_type_windows_server')"`

	Accounts []Account `gorm:"foreignKey:Host,Port"`
}
