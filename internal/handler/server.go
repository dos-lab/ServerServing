package handler

import (
	"ServerServing/api/format"
	SErr "ServerServing/err"
	models "ServerServing/internal/internal_models"
	"ServerServing/internal/service"
	"ServerServing/util"
	"github.com/gin-gonic/gin"
)

type ServerHandler struct{}

func GetServerHandler() *ServerHandler {
	return &ServerHandler{}
}

// Create
// @Summary 创建server。
// @Tags server
// @Produce json
// @Router /api/v1/servers/ [post]
// @Param serverCreateRequest body internal_models.ServerCreateRequest true "serverCreateRequest"
// @Success 200 {object} internal_models.ServerCreateResponse
func (ServerHandler) Create(c *gin.Context) (*format.JSONRespFormat, *SErr.APIErr) {
	req := &models.ServerCreateRequest{}
	e := c.ShouldBind(req)
	if e != nil {
		return nil, SErr.BadRequestErr
	}

	sessionsSvc := service.GetSessionsService()
	_, err := sessionsSvc.LoggedInAndIsAdmin(c)
	if err != nil {
		return nil, err
	}

	serversSvc := service.GetServersService()
	err = serversSvc.Create(c, req.Host, req.Port, req.OSType, req.AdminAccountName, req.AdminAccountPwd)
	if err != nil {
		return nil, err
	}

	return format.SimpleOKResp(&models.ServerCreateResponse{}), nil
}

// Delete
// @Summary 删除server。
// @Tags server
// @Produce json
// @Router /api/v1/servers/ [delete]
// @Param serverDeleteRequest body internal_models.ServerDeleteRequest true "serverDeleteRequest"
// @Success 200 {object} internal_models.ServerDeleteResponse
func (ServerHandler) Delete(c *gin.Context) (*format.JSONRespFormat, *SErr.APIErr) {
	req := &models.ServerDeleteRequest{}
	e := c.ShouldBind(req)
	if e != nil {
		return nil, SErr.BadRequestErr
	}

	sessionsSvc := service.GetSessionsService()
	_, err := sessionsSvc.LoggedInAndIsAdmin(c)
	if err != nil {
		return nil, err
	}

	serversSvc := service.GetServersService()
	err = serversSvc.Delete(c, req.Host, req.Port)
	if err != nil {
		return nil, err
	}

	return format.SimpleOKResp(&models.ServerCreateResponse{}), nil
}

// Info
// @Summary 查询server信息。
// @Tags server
// @Produce json
// @Router /api/v1/servers/{host}/{port} [get]
// @param host path string true "host"
// @param port path uint true "port"
// @Param serverInfoRequest query internal_models.ServerInfoRequest true "serverInfoRequest"
// @Success 200 {object} internal_models.ServerInfoResponse
func (ServerHandler) Info(c *gin.Context) (*format.JSONRespFormat, *SErr.APIErr) {
	host := c.Param("host")
	portStr := c.Param("port")
	port, err := util.ParseInt(portStr)
	if err != nil || port <= 0 {
		return nil, SErr.BadRequestErr.CustomMessageF("请求的Port不为整数！或者Port <= 0")
	}
	req := &models.ServerInfoRequest{}
	e := c.ShouldBindQuery(req)
	if e != nil {
		return nil, SErr.BadRequestErr
	}

	serversSvc := service.GetServersService()
	info, sErr := serversSvc.Info(c, host, uint(port), &req.LoadServerDetailArg)
	if sErr != nil {
		return nil, sErr
	}
	return format.SimpleOKResp(&models.ServerInfoResponse{
		ServerInfo: info,
	}), nil
}

// Infos
// @Summary 查询多个server信息。
// @Tags server
// @Produce json
// @Router /api/v1/servers/ [get]
// @Param serverInfosRequest query internal_models.ServerInfosRequest true "serverInfosRequest"
// @Success 200 {object} internal_models.ServerInfosResponse
func (ServerHandler) Infos(c *gin.Context) (*format.JSONRespFormat, *SErr.APIErr) {
	req := &models.ServerInfosRequest{}
	e := c.ShouldBindQuery(req)
	if e != nil {
		return nil, SErr.BadRequestErr
	}

	serversSvc := service.GetServersService()
	infos, totalCount, sErr := serversSvc.Infos(c, req.From, req.Size, &req.LoadServerDetailArg, req.Keyword)
	if sErr != nil {
		return nil, sErr
	}
	return format.SimpleOKResp(&models.ServerInfosResponse{
		ServerInfos: infos,
		TotalCount:  totalCount,
	}), nil
}
