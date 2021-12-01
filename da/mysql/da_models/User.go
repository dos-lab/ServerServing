package da_models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name  string `gorm:"index:idx_name,unique"`
	Pwd   string
	Admin bool
}
