package api

import (
	"ServerServing/api/format"
	"ServerServing/err"
	"ServerServing/internal/handler"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func Register(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	rg := r.Group("/api/v1")

	testRouter := rg.Group(prefixTest)
	usersRouter := rg.Group(prefixUser)
	sessionsRouter := rg.Group(prefixSession)
	serversRouter := rg.Group(prefixServer)
	serversAccountsRouter := rg.Group(prefixServerAccounts)

	testAPI := testAPI{}
	testRouter.GET("error_handler", format.Wrap(testAPI.testErrorHandler()))
	testRouter.GET("ping", format.Wrap(testAPI.ping()))

	usersAPI := usersAPI{}
	usersRouter.POST("", format.Wrap(usersAPI.create()))
	usersRouter.PUT(":id", format.Wrap(usersAPI.update()))
	usersRouter.GET(":id", format.Wrap(usersAPI.info()))
	usersRouter.GET("", format.Wrap(usersAPI.infos()))

	sessionsAPI := sessionsAPI{}
	sessionsRouter.POST("", format.Wrap(sessionsAPI.create()))
	sessionsRouter.DELETE("", format.Wrap(sessionsAPI.destroy()))
	sessionsRouter.GET("", format.Wrap(sessionsAPI.check()))

	serversAPI := serversAPI{}
	serversRouter.POST("", format.Wrap(serversAPI.create()))
	serversRouter.DELETE("", format.Wrap(serversAPI.delete()))
	serversRouter.GET(":host/:port", format.Wrap(serversAPI.info()))
	serversRouter.GET("", format.Wrap(serversAPI.infos()))

	serversAccountsAPI := serversAccountsAPI{}
	serversAccountsRouter.POST("", format.Wrap(serversAccountsAPI.create()))
	serversAccountsRouter.GET("/backupDir", format.Wrap(serversAccountsAPI.backupDir()))
	serversAccountsRouter.DELETE("", format.Wrap(serversAccountsAPI.delete()))
	serversAccountsRouter.PUT("", format.Wrap(serversAccountsAPI.update()))
}

const (
	prefixTest           = "test"
	prefixUser           = "users"
	prefixSession        = "sessions"
	prefixServer         = "servers"
	prefixServerAccounts = "servers/accounts/"
)

//type sourceCodeAPI struct{}
//
//func (sourceCodeAPI) localPatchSourceCode() api_format.JSONHandler {
//	return func(c *gin.Context) (*api_format.JSONRespFormat, *err.APIErr) {
//		return local.GetHandler().PatchSourceCode(c)
//	}
//}
//
//func (sourceCodeAPI) search() api_format.JSONHandler {
//	return func(c *gin.Context) (*api_format.JSONRespFormat, *err.APIErr) {
//		return search.GetHandler().SearchSourceCode(c)
//	}
//}

type usersAPI struct{}

func (usersAPI) create() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetUserHandler().Create(c)
	}
}

func (usersAPI) update() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetUserHandler().Update(c)
	}
}

func (usersAPI) info() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetUserHandler().Info(c)
	}
}

func (usersAPI) infos() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetUserHandler().Infos(c)
	}
}

type sessionsAPI struct{}

func (sessionsAPI) create() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetSessionsHandler().Create(c)
	}
}

func (sessionsAPI) destroy() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetSessionsHandler().Destroy(c)
	}
}

func (sessionsAPI) check() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetSessionsHandler().Check(c)
	}
}

type serversAPI struct{}

func (serversAPI) create() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetServerHandler().Create(c)
	}
}

func (serversAPI) delete() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetServerHandler().Delete(c)
	}
}

func (serversAPI) info() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetServerHandler().Info(c)
	}
}

func (serversAPI) infos() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetServerHandler().Infos(c)
	}
}

type serversAccountsAPI struct{}

func (serversAccountsAPI) create() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetServerAccountsHandler().Create(c)
	}
}

func (serversAccountsAPI) backupDir() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetServerAccountsHandler().BackupDirInfo(c)
	}
}

func (serversAccountsAPI) delete() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetServerAccountsHandler().Delete(c)
	}
}

func (serversAccountsAPI) update() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return handler.GetServerAccountsHandler().Update(c)
	}
}

type testAPI struct{}

// Ping
// @Summary ping
// @Tags test
// @Produce json
// @Router /api/v1/test/ping [get]
// @Success 200
// @Fail err.APIErr
func (testAPI) ping() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		return format.SimpleOKResp("OK"), nil
	}
}

// TestErrorHandler
// @Summary test error handler
// @Tags test
// @Produce json
// @Router /api/v1/test/error_handler [get]
// @Success 200
// @Fail err.APIErr
func (testAPI) testErrorHandler() format.JSONHandler {
	return func(c *gin.Context) (interface{}, *err.APIErr) {
		_ = c.AbortWithError(err.BadRequestErr.Status, err.BadRequestErr)
		return nil, nil
	}
}
