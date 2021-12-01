package main

import (
	"ServerServing/api"
	"ServerServing/config"
	"ServerServing/da/mysql"
	_ "ServerServing/docs"
	"ServerServing/middlewares"
	"fmt"
	"github.com/gin-gonic/gin"
)

// @title SoftwareWarehouse Web API
// @version 1.0
// @description This is a SoftwareWarehouse API server.
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath
func main() {
	config.InitConfig()
	mysql.InitMySQL()
	r := gin.Default()
	registerMiddleware(r)
	api.Register(r)
	addr := fmt.Sprintf("%s:%d", config.GetConfig().Host, config.GetConfig().Port)
	err := r.Run(addr) // listen and serve on 0.0.0.0:8080
	if err != nil {
		panic(err)
	}
}

func registerMiddleware(r *gin.Engine) {
	r.NoRoute(middlewares.NotFoundHandler())
	r.NoMethod(middlewares.NotFoundHandler())
	r.Use(middlewares.Cors())
	r.Use(middlewares.Recover())
	r.Use(middlewares.ErrHandler())
	r.Use(middlewares.RedisSession())
}
