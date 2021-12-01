package middlewares

import (
	"ServerServing/api/format"
	SErr "ServerServing/err"
	"github.com/gin-gonic/gin"
)

func ErrHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if length := len(c.Errors); length > 0 {
			e := c.Errors[length-1]
			err := e.Err
			if err != nil {
				var aErr *SErr.APIErr
				if e, ok := err.(*SErr.APIErr); ok {
					aErr = e
				} else if e, ok := err.(error); ok {
					aErr = SErr.ForbiddenErr.CustomMessage(e.Error())
				} else {
					aErr = SErr.InternalErr
				}
				// 记录一个错误的日志
				format.Err(c, aErr)
				return
			}
		}

	}
}

func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		format.Err(c, SErr.NotFoundErr)
	}
}
