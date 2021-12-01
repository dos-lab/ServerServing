package middlewares

import (
	"ServerServing/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

func RedisSession() gin.HandlerFunc {
	store, err := redis.NewStore(10, "tcp", config.GetConfig().RedisConfig.Addr, "", []byte(""))
	if err != nil {
		panic(err)
	}
	return sessions.Sessions("ServerServingSession", store)
}
