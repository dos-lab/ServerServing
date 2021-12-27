package format

import (
	SErr "ServerServing/err"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Wrap(handler interface{}) func(c *gin.Context) {
	var target func(c *gin.Context)
	switch handler.(type) {
	case JSONHandler:
		target = wrapJSONHandler(handler.(JSONHandler))
		break
	case NormalHandler:
		target = wrapNormalHandler(handler.(NormalHandler))
		break
	default:
		panic("Unsupported SearchSourceCode API type")
	}
	return target
}

func wrapJSONHandler(handler JSONHandler) func(c *gin.Context) {
	return func(c *gin.Context) {
		resp, e := handler(c)
		if len(c.Errors) > 0 {
			return
		}
		if e != nil {
			Err(c, e)
			return
		}
		c.JSON(http.StatusOK, SimpleOKResp(resp))
	}
}

func wrapNormalHandler(handler NormalHandler) func(c *gin.Context) {
	return func(c *gin.Context) {
		handler(c)
	}
}

func Err(c *gin.Context, e *SErr.APIErr) {
	c.JSON(http.StatusOK, NewJSONResp(e.Code, e.Message, nil))
}
