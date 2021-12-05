package dal

import (
	"ServerServing/da/mysql"
	daModels "ServerServing/da/mysql/da_models"
	SErr "ServerServing/err"
	"ServerServing/util"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type ServerDal struct{}

func GetServerDal() ServerDal {
	return ServerDal{}
}

func (s ServerDal) validServerInfo(server *daModels.Server) bool {
	return server.Host != "" &&
		server.Port != 0 &&
		len(server.Accounts) == 0 &&
		server.AdminAccountName != "" &&
		server.AdminAccountPwd != ""
}

// Create 创建一个新的Server。
func (s ServerDal) Create(server *daModels.Server) *SErr.APIErr {
	if !s.validServerInfo(server) {
		return SErr.InternalErr.CustomMessageF("创建Server的数据不合法！Server=[%v]", util.Pretty(server))
	}
	db := mysql.GetDB()
	res := db.Model(&daModels.Server{}).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(server)
	if res.Error != nil {
		return SErr.InternalErr
	}
	if res.RowsAffected == 0 {
		return SErr.InvalidParamErr.CustomMessageF("该Server Host=[%s] Port=[%d] 已存在，请检查参数！", server.Host, server.Port)
	}
	return nil
}

// Delete 删除一个服务器。
func (s ServerDal) Delete(Host string, Port uint) *SErr.APIErr {
	db := mysql.GetDB()
	res := db.Delete(&daModels.Server{Host: Host, Port: Port})
	if res.Error != nil {
		return SErr.InternalErr
	}
	if res.RowsAffected == 0 {
		return SErr.InvalidParamErr.CustomMessageF("要删除的Server Host=[%s] Port=[%d] 不存在，请检查参数！", Host, Port)
	}
	return nil
}

// Count 获取Server总数量。
func (s ServerDal) Count() (int64, *SErr.APIErr) {
	var count int64
	db := mysql.GetDB()
	res := db.Model(&daModels.Server{}).Count(&count)
	if res.Error != nil {
		return 0, SErr.InternalErr.CustomMessageF("查询Server数量时出错，出错信息为：[%s]", res.Error.Error())
	}
	return count, nil
}

// Get 获取一个Server的信息。需要指定Host和Port
func (s ServerDal) Get(Host string, Port uint, withAccounts bool) (*daModels.Server, *SErr.APIErr) {
	db := mysql.GetDB()
	server := &daModels.Server{}
	var res *gorm.DB
	if withAccounts {
		res = db.Model(&daModels.Server{}).Preload("Accounts", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).Where(&daModels.Server{Host: Host, Port: Port}).First(server)
	} else {
		res = db.Model(&daModels.Server{}).Where(&daModels.Server{Host: Host, Port: Port}).First(server)
	}
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		// 找不到该Host和Port对应的Server
		return nil, SErr.InvalidParamErr.CustomMessageF("找不到该Server，参数为Host=[%s], Port=[%d]", Host, Port)
	}
	if res.Error != nil {
		return nil, SErr.InternalErr.CustomMessageF("查询Server时出错！出错信息为：[%s]", res.Error.Error())
	}
	return server, nil
}

// List 获取Server列表。
func (s ServerDal) List(from, size uint, withAccounts bool) ([]*daModels.Server, uint, *SErr.APIErr) {
	log.Printf("Server List, from=[%d], size=[%d]", from, size)
	var servers []*daModels.Server
	count, err := s.Count()
	if err != nil {
		return nil, 0, err
	}
	db := mysql.GetDB()
	var res *gorm.DB
	if withAccounts {
		res = db.Model(&daModels.Server{}).Preload("Accounts", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).Order("created_at desc").Offset(int(from)).Limit(int(size)).Find(&servers)
	} else {
		res = db.Model(&daModels.Server{}).Order("created_at desc").Offset(int(from)).Limit(int(size)).Find(&servers)
	}
	if res.Error != nil {
		return nil, 0, SErr.InternalErr.CustomMessageF("查询Server列表时出错！出错信息为：[%s]", res.Error.Error())
	}
	return servers, uint(count), nil
}

// SearchByHostAndAdmin 指定一个keyword，同时针对Host和Admin做搜索。
func (s ServerDal) SearchByHostAndAdmin(from, size uint, keyword string, withAccounts bool) ([]*daModels.Server, uint, *SErr.APIErr) {
	log.Printf("Servers SearchByHostAndAdmin, keyword=[%s], from=[%d], size=[%d], withAccounts=[%v]", keyword, from, size, withAccounts)
	var servers []*daModels.Server
	var count int64
	db := mysql.GetDB()
	query := "Host LIKE ? or AdminAccountName LIKE ?"
	args := []interface{}{"%" + keyword + "%", "%" + keyword + "%"}
	res := db.Model(&daModels.Server{}).Where(query, args...).Count(&count)
	if res.Error != nil {
		return nil, 0, SErr.InternalErr.CustomMessageF("搜索Server时，查询总长度失败！出错信息为：[%s]", res.Error.Error())
	}
	if withAccounts {
		// Preload with unscoped accounts
		res = db.Model(&daModels.Server{}).Preload("Accounts", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).Where(query, args...).Offset(int(from)).Limit(int(size)).Find(&servers)
	} else {
		res = db.Model(&daModels.Server{}).Where(query, args...).Offset(int(from)).Limit(int(size)).Find(&servers)
	}
	if res.Error != nil {
		return nil, 0, SErr.InternalErr.CustomMessageF("搜索Server时出错！出错信息为：[%s]", res.Error.Error())
	}
	return servers, uint(count), nil
}

// Update 更新数据库信息，无法保证更新的目标是否相同。可以在上层通过redis做分布式锁做并发更新保障。（在mysql做并发控制需要使用事务，代价比较高）
func (s ServerDal) Update(Host string, Port uint, server *daModels.Server) *SErr.APIErr {
	if !s.validServerInfo(server) {
		return SErr.InvalidParamErr.CustomMessageF("更新Server数据时，参数不合法！参数为：[%s]", util.Pretty(server))
	}
	db := mysql.GetDB()
	res := db.Model(&daModels.Server{}).Where(&daModels.Server{Host: Host, Port: Port}).First(server)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return SErr.InvalidParamErr.CustomMessageF("更新Server数据时，更新的目标数据不存在！目标Server信息为：Host=[%s], Port=[%d]", Host, Port)
	}
	res = db.Model(&daModels.Server{}).Where(&daModels.Server{Host: Host, Port: Port}).Updates(server)
	if res.Error != nil {
		return SErr.InternalErr.CustomMessageF("更新Server数据出错，出错信息为：[%s]", res.Error.Error())
	}
	return nil
}
