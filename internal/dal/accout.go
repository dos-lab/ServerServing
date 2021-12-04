package dal

import (
	"ServerServing/da/mysql"
	daModels "ServerServing/da/mysql/da_models"
	SErr "ServerServing/err"
	"gorm.io/gorm/clause"
	"log"
)

type AccountDal struct {}

func GetAccountDal() AccountDal {
	return AccountDal{}
}

func (a AccountDal) Count(Host string, Port uint) (int, *SErr.APIErr) {
	var count int64
	db := mysql.GetDB()
	res := db.Model(&daModels.Account{}).Where(&daModels.Account{Host: Host, Port: Port}).Count(&count)
	if res.Error != nil {
		return 0, SErr.InternalErr.CustomMessageF("查询Account数量时出错，出错信息为：[%s]", res.Error.Error())
	}
	return int(count), nil
}

func (a AccountDal) List(Host string, Port, from, size uint) ([]*daModels.Account, int, *SErr.APIErr) {
	log.Printf("Account List, Host=[%s], Port=[%d], from=[%d], size=[%d]", Host, Port, from, size)
	var accounts []*daModels.Account
	count, err := a.Count(Host, Port)
	if err != nil {
		return nil, 0, err
	}
	db := mysql.GetDB()
	res := db.Model(&daModels.Account{}).Where(&daModels.Account{Host: Host, Port: Port}).Order("CreatedAt desc").Offset(int(from)).Limit(int(size)).Find(&accounts)
	if res.Error != nil {
		return nil, 0, SErr.InternalErr.CustomMessageF("查询Account列表时出错！出错信息为：[%s]", res.Error.Error())
	}
	return accounts, count, nil
}

func (a AccountDal) Upsert(accounts []*daModels.Account) *SErr.APIErr {
	db := mysql.GetDB()
	res := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "Host"}, {Name: "Port"}, {Name: "Name"}},
		DoUpdates: clause.AssignmentColumns([]string{"Pwd"}),
	}).Create(&accounts)
	if res.Error != nil {
		return SErr.InternalErr.CustomMessageF("更新/插入新的账户数据时出错，错误信息为：[%s]", res.Error.Error())
	}
	return nil
}