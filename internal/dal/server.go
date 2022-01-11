package dal

import (
	"ServerServing/da/mysql"
	daModels "ServerServing/da/mysql/da_models"
	SErr "ServerServing/err"
	"ServerServing/util"
	"errors"
	"gorm.io/gorm"
	"log"
)

type ServerDal struct{}

func GetServerDal() ServerDal {
	return ServerDal{}
}

func (s ServerDal) validServerInfo(server *daModels.Server) bool {
	return server.Name != "" &&
		server.Host != "" &&
		server.Port != 0 &&
		len(server.Accounts) == 0 &&
		server.AdminAccountName != "" &&
		server.AdminAccountPwd != ""
}

// Create 创建一个新的Server。
func (s ServerDal) Create(server *daModels.Server) *SErr.APIErr {
	if !s.validServerInfo(server) {
		return SErr.InvalidParamErr.CustomMessageF("创建Server的数据不合法！Server=[%v]", util.Pretty(server))
	}
	db := mysql.GetDB()
	var sErr *SErr.APIErr
	_ = db.Transaction(func(tx *gorm.DB) error {
		tmpServer := &daModels.Server{}
		res := tx.Model(&daModels.Server{}).Where("Host = ? and Port = ? or Name = ?", server.Host, server.Port, server.Name).First(tmpServer)
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			if res.Error != nil {
				sErr = SErr.InternalErr.CustomMessageF("创建服务器时，查询已有信息出错！数据库出错=[%v]", res.Error)
				return res.Error
			}
			// 查找到了已有的Host和Port的数据。返回错误。
			if tmpServer.Host == server.Host && tmpServer.Port == server.Port {
				sErr = SErr.InvalidParamErr.CustomMessage("该服务器的Host与Port已存在！")
				return errors.New("重复的服务器Host与Port")
			}
			if tmpServer.Name == server.Name {
				sErr = SErr.InvalidParamErr.CustomMessage("该服务器名称已存在！")
				return errors.New("重复的服务器名称")
			}
		}
		res = db.Model(&daModels.Server{}).Create(server)
		if res.Error != nil {
			sErr = SErr.InternalErr.CustomMessageF("创建服务器失败！出错信息=[%v]", res.Error)
			return res.Error
		}
		return nil
	})
	//res := db.Model(&daModels.Server{}).Clauses(clause.OnConflict{
	//	UpdateAll: true,
	//}).Create(server)
	return sErr
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

// Search 指定一个keyword，同时针对Host和Admin做搜索。
func (s ServerDal) Search(from, size uint, keyword string, withAccounts bool) ([]*daModels.Server, uint, *SErr.APIErr) {
	log.Printf("Servers Search, keyword=[%s], from=[%d], size=[%d], withAccounts=[%v]", keyword, from, size, withAccounts)
	var servers []*daModels.Server
	var count int64
	db := mysql.GetDB()
	query := "Name LIKE ? or Host LIKE ? or admin_account_name LIKE ?"
	likeParam := "%" + keyword + "%"
	args := []interface{}{likeParam, likeParam, likeParam}
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
	db := mysql.GetDB()
	var sErr *SErr.APIErr
	_ = db.Transaction(func(tx *gorm.DB) error {
		tmpServer := &daModels.Server{}
		res := db.Model(&daModels.Server{}).Where(&daModels.Server{Host: Host, Port: Port}).First(tmpServer)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			sErr = SErr.InvalidParamErr.CustomMessageF("更新Server数据时，更新的目标的服务器[%s]，Port=[%d]不存在！", Host, Port)
			return res.Error
		}
		res = db.Model(&daModels.Server{}).Where(&daModels.Server{Name: server.Name}).First(tmpServer)
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) && res.Error == nil {
			sErr = SErr.InvalidParamErr.CustomMessage("更新Server数据时，服务器名称重复！")
			return errors.New("服务器名称重复")
		}
		res = db.Model(&daModels.Server{}).Where(&daModels.Server{Host: Host, Port: Port}).Select("name", "description", "admin_account_name", "admin_account_pwd").Updates(server)
		if res.Error != nil {
			sErr = SErr.InternalErr.CustomMessageF("更新Server数据出错，出错信息为：[%s]", res.Error.Error())
			return res.Error
		}
		return nil
	})
	return sErr
}
