package format

import (
	"ServerServing/err"
	"github.com/gin-gonic/gin"
)

type JSONHandler func(*gin.Context) (interface{}, *err.APIErr)
type NormalHandler func(*gin.Context)
