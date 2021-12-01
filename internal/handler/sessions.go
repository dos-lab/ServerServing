package handler

import (
	"ServerServing/api/format"
	SErr "ServerServing/err"
	"ServerServing/internal/service"
	"github.com/gin-gonic/gin"
)

type SessionsHandler struct{}

func GetSessionsHandler() *SessionsHandler {
	return &SessionsHandler{}
}

// Create
// @Summary 创建session。（登录）
// @Tags session
// @Produce json
// @Router /api/v1/sessions/ [post]
// @Param createRequest body api_models.SessionsCreateRequest true "createRequest"
// @Success 200 {object} api_models.SessionsCreateResponse
func (SessionsHandler) Create(c *gin.Context) (*format.JSONRespFormat, *SErr.APIErr) {
	req := &models.SessionsCreateRequest{}
	err := c.ShouldBind(req)
	if err != nil {
		return nil, SErr.BadRequestErr
	}
	sErr := service.GetSessionsService().Create(c, req.Name, req.Pwd)
	if sErr != nil {
		return nil, sErr
	}
	return format.SimpleOKResp(&models.SessionsCreateResponse{}), nil
}

// Destroy
// @Summary 退出session。（退出登录）
// @Tags session
// @Produce json
// @Router /api/v1/sessions/ [delete]
// @Param destroyRequest body api_models.SessionsDestroyRequest true "destroyRequest"
// @Success 200 {object} api_models.SessionsDestroyResponse
func (SessionsHandler) Destroy(c *gin.Context) (*format.JSONRespFormat, *SErr.APIErr) {
	service.GetSessionsService().Destroy(c)
	return format.SimpleOKResp(&models.SessionsDestroyResponse{}), nil
}

// Check
// @Summary 检查登录状态。
// @Tags session
// @Produce json
// @Router /api/v1/sessions/ [get]
// @Success 200 {object} api_models.SessionsCheckResponse
func (SessionsHandler) Check(c *gin.Context) (*format.JSONRespFormat, *SErr.APIErr) {
	userID, sErr := service.GetSessionsService().GetUserID(c)
	if sErr != nil {
		return nil, sErr
	}
	return format.SimpleOKResp(&models.SessionsCheckResponse{
		UserID: userID,
	}), nil
}
