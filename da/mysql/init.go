package mysql

import (
	"ServerServing/config"
	"ServerServing/da/mysql/da_models"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var db *gorm.DB

func InitMySQL() {
	confParam := config.GetConfig().MySqlConfig
	var err error
	DSN := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", confParam.UserName, confParam.Pwd, confParam.Addr, confParam.DBName)
	log.Printf("MySQL Connection Establishing... DSN=[%s]", DSN)
	db, err = gorm.Open(mysql.Open(DSN))
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&da_models.User{})
	if err != nil {
		panic(err)
	}
}

func GetDB() *gorm.DB {
	return db
}
