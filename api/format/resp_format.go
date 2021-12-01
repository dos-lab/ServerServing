package format

import "net/http"

type JSONRespFormat struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SimpleOKResp(data interface{}) *JSONRespFormat {
	return &JSONRespFormat{
		Status:  http.StatusOK,
		Message: "success",
		Data:    data,
	}
}

func NewJSONResp(statusCode int, msg string, data interface{}) *JSONRespFormat {
	return &JSONRespFormat{
		Status:  statusCode,
		Message: msg,
		Data:    data,
	}
}
