package middlewares

import (
	"ServerServing/api/format"
	SErr "ServerServing/err"
	"github.com/gin-gonic/gin"
	"log"
	"runtime/debug"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic: %v\n", r)
				debug.PrintStack()
				format.Err(c, SErr.InternalErr)
				c.Abort()
			}
		}()
		c.Next()
	}
}
