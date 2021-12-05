package service

import (
	SErr "ServerServing/err"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SessionsService struct{}

func GetSessionsService() *SessionsService {
	return &SessionsService{}
}

var userIDSessionKey = "userID"

func (*SessionsService) Create(c *gin.Context, name, pwd string) *SErr.APIErr {
	// 首先检查name和pwd正确性。
	userService := GetUsersService()
	u, sErr := userService.CheckPwd(c, name, pwd)
	if sErr != nil {
		return sErr
	}
	session := sessions.Default(c)
	session.Set(userIDSessionKey, u.ID)
	err := session.Save()
	if err != nil {
		return SErr.InternalErr
	}
	return nil
}

func (*SessionsService) GetUserID(c *gin.Context) (int, *SErr.APIErr) {
	session := sessions.Default(c)
	intF := session.Get(userIDSessionKey)
	if intF == nil {
		return 0, nil
	}
	userID := intF.(uint)
	if userID == 0 {
		return 0, SErr.NeedLoginErr
	}
	return int(userID), nil
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
