package service

import (
	"ServerServing/da/mysql/da_models"
	SErr "ServerServing/err"
	"ServerServing/internal/dal"
	"ServerServing/internal/internal_models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
)

type UsersService struct{}

func GetUsersService() *UsersService {
	return &UsersService{}
}

func (s *UsersService) Create(c *gin.Context, name string, pwd string) (string, *SErr.APIErr) {
	user := &da_models.User{
		Name: name,
		Pwd:  pwd,
	}
	sErr := dal.GetUserDal().Create(user)
	if sErr != nil {
		return "", sErr
	}
	token, sErr := GetSessionsService().Create(c, name, pwd)
	if sErr != nil {
		return "", sErr
	}
	return token, nil
}

func (s *UsersService) Info(c *gin.Context, userID int) (*da_models.User, *SErr.APIErr) {
	user, sErr := dal.GetUserDal().GetByID(userID)
	if sErr != nil {
		return nil, sErr
	}
	return user, sErr
}

func (s *UsersService) CheckPwd(c *gin.Context, name string, pwd string) (*da_models.User, *SErr.APIErr) {
	user, sErr := dal.GetUserDal().GetByName(name)
	if sErr != nil {
		return nil, sErr
	}
	if user.Pwd != pwd {
		return nil, SErr.WrongPwdErr
	}
	return user, sErr
}

func (s *UsersService) Infos(c *gin.Context, from, size int, searchKeyword *string) ([]*da_models.User, int, *SErr.APIErr) {
	var users []*da_models.User
	var totalCount int
	var sErr *SErr.APIErr
	if searchKeyword == nil || *searchKeyword == "" {
		users, totalCount, sErr = dal.GetUserDal().List(from, size)
	} else {
		users, totalCount, sErr = dal.GetUserDal().SearchByName(*searchKeyword, from, size)
	}
	return users, totalCount, sErr
}

func (s *UsersService) Update(c *gin.Context, targetID int, updateReq *internal_models.UsersUpdateRequest) *SErr.APIErr {
	userID, sErr := GetSessionsService().GetUserID(c)
	if sErr != nil {
		return SErr.NeedLoginErr.CustomMessageF("更新用户信息前，必须要登录。")
	}
	updateReqStr, _ := json.Marshal(updateReq)
	// 只有管理员可以更新其他人的数据，只有本人可以修改本人的数据
	userInfo, sErr := s.Info(c, userID)
	userInfoStr, _ := json.Marshal(userInfo)
	log.Printf("UsersService Update userID=[%d], targetID=[%d], updateReq=[%s], userInfo=[%s]", userID, targetID, updateReqStr, userInfoStr)
	if sErr != nil {
		return sErr
	}
	if !userInfo.Admin && userID != targetID {
		return SErr.AdminOnlyActionErr
	}
	if !userInfo.Admin && updateReq.Admin != nil && userID == targetID {
		return SErr.AdminOnlyActionErr
	}
	return dal.GetUserDal().Update(targetID, updateReq)
}

func (s *UsersService) IsAdmin(c *gin.Context, targetID int) (bool, *SErr.APIErr) {
	userInfo, err := s.Info(c, targetID)
	if err != nil {
		return false, err
	}
	return userInfo.Admin, nil
}
