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
}

const (
	prefixTest    = "test"
	prefixUser    = "users"
	prefixSession = "sessions"
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
	return func(c *gin.Context) (*format.JSONRespFormat, *err.APIErr) {
		return handler.GetUserHandler().Create(c)
	}
}

func (usersAPI) update() format.JSONHandler {
	return func(c *gin.Context) (*format.JSONRespFormat, *err.APIErr) {
		return handler.GetUserHandler().Update(c)
	}
}

func (usersAPI) info() format.JSONHandler {
	return func(c *gin.Context) (*format.JSONRespFormat, *err.APIErr) {
		return handler.GetUserHandler().Info(c)
	}
}

func (usersAPI) infos() format.JSONHandler {
	return func(c *gin.Context) (*format.JSONRespFormat, *err.APIErr) {
		return handler.GetUserHandler().Infos(c)
	}
}

type sessionsAPI struct{}

func (sessionsAPI) create() format.JSONHandler {
	return func(c *gin.Context) (*format.JSONRespFormat, *err.APIErr) {
		return handler.GetSessionsHandler().Create(c)
	}
}

func (sessionsAPI) destroy() format.JSONHandler {
	return func(c *gin.Context) (*format.JSONRespFormat, *err.APIErr) {
		return handler.GetSessionsHandler().Destroy(c)
	}
}

func (sessionsAPI) check() format.JSONHandler {
	return func(c *gin.Context) (*format.JSONRespFormat, *err.APIErr) {
		return handler.GetSessionsHandler().Check(c)
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
	return func(c *gin.Context) (*format.JSONRespFormat, *err.APIErr) {
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
	return func(c *gin.Context) (*format.JSONRespFormat, *err.APIErr) {
		_ = c.AbortWithError(err.BadRequestErr.Status, err.BadRequestErr)
		return nil, nil
	}
}
