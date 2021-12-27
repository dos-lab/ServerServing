package format

import "ServerServing/err"

type JSONRespFormat struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SimpleOKResp(data interface{}) *JSONRespFormat {
	return &JSONRespFormat{
		Code:    err.CodeOK,
		Message: "success",
		Data:    data,
	}
}

func NewJSONResp(statusCode int, msg string, data interface{}) *JSONRespFormat {
	return &JSONRespFormat{
		Code:    statusCode,
		Message: msg,
		Data:    data,
	}
}
