package handler

import (
	SErr "ServerServing/err"
	models "ServerServing/internal/internal_models"
	"ServerServing/internal/service"
	"ServerServing/util"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
)

type ServerHandler struct{}

func GetServerHandler() *ServerHandler {
	return &ServerHandler{}
}

// ConnectionTest
// @Summary 测试连通性
// @Tags server
// @Produce json
// @Router /api/v1/servers/connections/{host}/{port} [get]
// @param host path string true "host"
// @param port path uint true "port"
// @Param serverConnectionTestRequest query internal_models.ServerConnectionTestRequest true "serverConnectionTestRequest"
// @Success 200 {object} internal_models.ServerConnectionTestResponse
func (ServerHandler) ConnectionTest(c *gin.Context) (interface{}, *SErr.APIErr) {
	host := c.Param("host")
	portStr := c.Param("port")
	port, err := util.ParseInt(portStr)
	if err != nil || port <= 0 {
		return nil, SErr.BadRequestErr.CustomMessageF("请求的Port不为整数！或者Port <= 0")
	}
	req := &models.ServerConnectionTestRequest{}
	reqStr, _ := json.Marshal(req)
	log.Printf("ServerHandler ConnectionTest reqStr=[%s]", reqStr)
	e := c.ShouldBindQuery(req)
	if e != nil {
		return nil, SErr.BadRequestErr
	}

	serversSvc := service.GetServersService()
	sErr := serversSvc.ConnectionTest(c, host, uint(port), req.OSType, req.AccountName, req.AccountPwd)
	if sErr != nil {
		return &models.ServerConnectionTestResponse{
			Connected: false,
			Cause:     sErr.Message,
		}, nil
	}
	return &models.ServerConnectionTestResponse{Connected: true}, nil
}

// Create
// @Summary 创建server。
// @Tags server
// @Produce json
// @Router /api/v1/servers/ [post]
// @Param serverCreateRequest body internal_models.ServerCreateRequest true "serverCreateRequest"
// @Success 200 {object} internal_models.ServerCreateResponse
func (ServerHandler) Create(c *gin.Context) (interface{}, *SErr.APIErr) {
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

	return &models.ServerCreateResponse{}, nil
}

// Delete
// @Summary 删除server。
// @Tags server
// @Produce json
// @Router /api/v1/servers/ [delete]
// @Param serverDeleteRequest body internal_models.ServerDeleteRequest true "serverDeleteRequest"
// @Success 200 {object} internal_models.ServerDeleteResponse
func (ServerHandler) Delete(c *gin.Context) (interface{}, *SErr.APIErr) {
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

	return &models.ServerCreateResponse{}, nil
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
func (ServerHandler) Info(c *gin.Context) (interface{}, *SErr.APIErr) {
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
	return &models.ServerInfoResponse{
		ServerInfo: info,
	}, nil
}

// Infos
// @Summary 查询多个server信息。
// @Tags server
// @Produce json
// @Router /api/v1/servers/ [get]
// @Param serverInfosRequest query internal_models.ServerInfosRequest true "serverInfosRequest"
// @Success 200 {object} internal_models.ServerInfosResponse
func (ServerHandler) Infos(c *gin.Context) (interface{}, *SErr.APIErr) {
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
	return &models.ServerInfosResponse{
		Infos:      infos,
		TotalCount: totalCount,
	}, nil
}
