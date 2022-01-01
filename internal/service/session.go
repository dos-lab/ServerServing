package service

import (
	daModels "ServerServing/da/mysql/da_models"
	SErr "ServerServing/err"
	"encoding/base64"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"strconv"
)

type SessionsService struct{}

func GetSessionsService() *SessionsService {
	return &SessionsService{}
}

// var userIDSessionKey = "userID"

const tokenKey = "X-Token"

func (*SessionsService) GenToken(user *daModels.User) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(user.ID))))
}

func (*SessionsService) ParseToken(token string) (uint, *SErr.APIErr) {
	bytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0, SErr.BadRequestErr.CustomMessageF("解析Token出错，err=[%s]", err)
	}
	uid, err := strconv.Atoi(string(bytes))
	if err != nil {
		return 0, SErr.BadRequestErr.CustomMessageF("解析Token时，转换整数出错，err=[%s]", err)
	}
	if uid < 0 {
		return 0, SErr.BadRequestErr.CustomMessageF("解析Token时，转换整数小于0，err=[%s]", err)
	}
	return uint(uid), nil
}

func (s *SessionsService) Create(c *gin.Context, name, pwd string) (string, *SErr.APIErr) {
	// 首先检查name和pwd正确性。
	userService := GetUsersService()
	u, sErr := userService.CheckPwd(c, name, pwd)
	if sErr != nil {
		return "", sErr
	}
	token := s.GenToken(u)
	//session := sessions.Default(c)
	//session.Set(userIDSessionKey, u.ID)
	//err := session.Save()
	//if err != nil {
	//	return SErr.InternalErr
	//}
	return token, nil
}

func (s *SessionsService) GetUserID(c *gin.Context) (int, *SErr.APIErr) {
	tokenHeader := c.GetHeader(tokenKey)
	uid, err := s.ParseToken(tokenHeader)
	if err != nil {
		return 0, err
	}
	//session := sessions.Default(c)
	//intF := session.Get(userIDSessionKey)
	//if intF == nil {
	//	return 0, nil
	//}
	//userID := intF.(uint)
	if uid == 0 {
		return 0, SErr.NeedLoginErr
	}
	return int(uid), nil
}

func (s *SessionsService) LoggedInAndIsAdmin(c *gin.Context) (int, *SErr.APIErr) {
	userID, err := s.GetUserID(c)
	if err != nil {
		return 0, SErr.NeedLoginErr
	}
	usersSvc := GetUsersService()
	isAdmin, err := usersSvc.IsAdmin(c, userID)
	if err != nil {
		return 0, err
	}
	if !isAdmin {
		return 0, SErr.AdminOnlyActionErr
	}
	return userID, nil
}

func (*SessionsService) Destroy(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
}
