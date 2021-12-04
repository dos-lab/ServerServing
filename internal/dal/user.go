package dal

import (
	"ServerServing/da/mysql"
	daModels "ServerServing/da/mysql/da_models"
	SErr "ServerServing/err"
	"ServerServing/internal/internal_models"
	"gorm.io/gorm"
	"log"
)

type UserDal struct{}

func GetDal() UserDal {
	return UserDal{}
}

func (UserDal) Create(user *daModels.User) *SErr.APIErr {
	db := mysql.GetDB()
	res := db.Where(&daModels.User{Name: user.Name}).FirstOrCreate(&user)
	if res.Error != nil {
		return SErr.InternalErr
	}
	if res.RowsAffected == 0 {
		return SErr.InvalidParamErr.CustomMessage("用户名已存在")
	}
	return nil
}

func (UserDal) GetByID(userID int) (*daModels.User, *SErr.APIErr) {
	db := mysql.GetDB()
	user := &daModels.User{
		Model: gorm.Model{
			ID: uint(userID),
		},
	}
	res := db.First(user)
	if res.Error != nil {
		return nil, SErr.InternalErr
	}
	if res.RowsAffected != 1 {
		return nil, SErr.BadRequestErr.CustomMessageF("用户不存在，userID=[%d]", userID)
	}
	return user, nil
}

func (UserDal) GetByName(name string) (*daModels.User, *SErr.APIErr) {
	db := mysql.GetDB()
	user := &daModels.User{}
	res := db.Model(&daModels.User{}).Where("name = ?", name).First(user)
	if res.Error != nil {
		return nil, SErr.InternalErr
	}
	if res.RowsAffected != 1 {
		return nil, SErr.BadRequestErr.CustomMessageF("用户不存在，userName=[%s]", name)
	}
	return user, nil
}

func (UserDal) List(from, size int) ([]*daModels.User, int, *SErr.APIErr) {
	log.Printf("Users List, from=[%d], size=[%d]", from, size)
	var users []*daModels.User
	var count int64
	db := mysql.GetDB()
	res := db.Model(&daModels.User{}).Count(&count)
	if res.Error != nil {
		return nil, 0, SErr.InternalErr.CustomMessage(res.Error.Error())
	}
	res = db.Model(&daModels.User{}).Order("CreatedAt desc").Offset(from).Limit(size).Find(&users)
	if res.Error != nil {
		return nil, 0, SErr.InternalErr.CustomMessage(res.Error.Error())
	}
	return users, int(count), nil
}

func (UserDal) SearchByName(keyword string, from, size int) ([]*daModels.User, int, *SErr.APIErr) {
	log.Printf("Users SearchByName, keyword=[%s], from=[%d], size=[%d]", keyword, from, size)
	var users []*daModels.User
	var count int64
	db := mysql.GetDB()
	res := db.Model(&daModels.User{}).Where("name LIKE ?", "%"+keyword+"%").Order("CreatedAt desc").Count(&count)
	if res.Error != nil {
		return nil, 0, SErr.InternalErr.CustomMessage(res.Error.Error())
	}
	res = db.Model(&daModels.User{}).Where("name LIKE ?", "%"+keyword+"%").Order("CreatedAt desc").Offset(from).Limit(size).Find(&users)
	if res.Error != nil {
		return nil, 0, SErr.InternalErr.CustomMessage(res.Error.Error())
	}
	return users, int(count), nil
}

func (UserDal) Count() (int64, *SErr.APIErr) {
	var count int64
	db := mysql.GetDB()
	res := db.Model(&daModels.User{}).Count(&count)
	if res.Error != nil {
		return 0, SErr.InternalErr.CustomMessage(res.Error.Error())
	}
	return count, nil
}

func (UserDal) Update(targetID int, updateReq *internal_models.UsersUpdateRequest) *SErr.APIErr {
	db := mysql.GetDB()
	updates := map[string]interface{}{
		"name":  updateReq.Name,
		"pwd":   updateReq.Pwd,
		"admin": updateReq.Admin,
	}
	omits := make([]string, 0, len(updates))
	//for k, v := range updates {
	//	if v == nil {
	//		omits = append(omits, k)
	//	}
	//}
	res := db.Model(&daModels.User{}).Where("id = ?", targetID).Omit(omits...).Updates(updates)
	if res.Error != nil {
		return SErr.InternalErr.CustomMessage(res.Error.Error())
	}
	return nil
}
