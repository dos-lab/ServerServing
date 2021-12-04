package da_models

import (
	"gorm.io/gorm"
	"time"
)

type OSType string

const (
	OSTypeLinux OSType = "os_type_linux"
	OSTypeWindowsServer OSType = "os_type_windows_server"
)

type Server struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Host string `gorm:"primaryKey;size:20"`
	Port uint 	`gorm:"primaryKey"`
	AdminAccountName string `gorm:"index;not null;size:50"`
	AdminAccountPwd string `gorm:"not null;size:50"`
	OSType OSType `json:"os_type" gorm:"not null" sql:"type:ENUM('os_type_linux', 'os_type_windows_server')"`

	Accounts []Account `gorm:"foreignKey:Host,Port"`
}