package err

import (
	"fmt"
	"net/http"
)

var (
	InternalErr = &APIErr{
		Message: "服务器内部错误！",
		Status:  http.StatusInternalServerError,
		Stable:  false,
	}
	NotFoundErr = &APIErr{
		Message: "未找到该API",
		Status:  http.StatusNotFound,
		Stable:  true,
	}
	BadRequestErr = &APIErr{
		Message: "错误的请求格式",
		Status:  http.StatusBadRequest,
		Stable:  true,
	}
	AdminOnlyActionErr = &APIErr{
		Message: "该行为需要管理员权限！",
		Status:  http.StatusOK,
		Stable:  true,
	}
	NeedLoginErr = &APIErr{
		Message: "需要登录！",
		Status:  http.StatusOK,
		Stable:  true,
	}
	InvalidParamErr = &APIErr{
		Message: "请求参数有误",
		Status:  http.StatusOK,
		Stable:  true,
	}
	ForbiddenErr = &APIErr{
		Message: "禁止访问",
		Status:  http.StatusForbidden,
		Stable:  true,
	}
	WrongPwdErr = &APIErr{
		Message: "密码有误！",
		Status:  http.StatusOK,
		Stable:  true,
	}
	SSHConnectionErr = &APIErr{
		Message: "与服务器建立SSH连接失败！",
		Status:  http.StatusOK,
		Stable:  true,
	}
	BackupDirNotExists = &APIErr{
		Message: "备份用户文件夹时，该用户的home目录文件夹不存在！",
		Status:  http.StatusOK,
		Stable:  true,
	}
	BackupTargetDirAlreadyExists = &APIErr{
		Message: "备份用户文件夹时，目标的文件夹已经存在！",
		Status:  http.StatusOK,
		Stable:  true,
	}
	RecoverHomeDirAlreadyExists = &APIErr{
		Message: "恢复备份的用户文件夹时，该用户的home目录文件夹已经存在！",
		Status:  http.StatusOK,
		Stable:  true,
	}
	CreateAccountNameAlreadyExists = &APIErr{
		Message: "创建账户时，该账户名已经存在！",
		Status:  http.StatusOK,
		Stable:  true,
	}
	UpdateAccountNameNotExists = &APIErr{
		Message: "更新账户信息时，该账户名不存在！",
		Status:  http.StatusOK,
		Stable:  true,
	}
	DeleteAccountNameNotExists = &APIErr{
		Message: "删除账户时，要删除的账户名不存在！",
		Status:  http.StatusOK,
		Stable:  true,
	}
)

type APIErr struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Stable  bool   `json:"stable"`
}

func (A *APIErr) Error() string {
	return A.Message
}

func (A *APIErr) CustomMessage(message string) *APIErr {
	return &APIErr{
		Message: message,
		Status:  A.Status,
		Stable:  A.Stable,
	}
}

func (A *APIErr) CustomMessageF(msg string, formatter ...interface{}) *APIErr {
	return &APIErr{
		Message: fmt.Sprintf(msg, formatter...),
		Status:  A.Status,
		Stable:  A.Stable,
	}
}
