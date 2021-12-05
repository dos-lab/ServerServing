package da_models

import (
	"gorm.io/gorm"
	"time"
)

// Account 表示某个服务器的账户
// 明明在Server中可以直接查询到该Server的数据，那么为何还需要在MySQL中再存一份呢？
// 这是因为在服务器中不能直接查询到账户的密码明文（只存储加密后的值）。
// 所以在MySQL中，存储那些在Server中不能够获取到的数据即可。
type Account struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name string `gorm:"primaryKey;index:idx_accounts_host_port_name,priority:3;size:50"`
	Pwd  string `gorm:"size:50"`

	Host string `gorm:"primaryKey;index:idx_accounts_host_port_name,priority:1;not null;size:20"`
	Port uint   `gorm:"primaryKey;index:idx_accounts_host_port_name,priority:2;not null"`

	Server Server `gorm:"foreignKey:Host,Port"`
}
