// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/sessions/": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "session"
                ],
                "summary": "检查登录状态。",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api_models.SessionsCheckResponse"
                        }
                    }
                }
            },
            "post": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "session"
                ],
                "summary": "创建session。（登录）",
                "parameters": [
                    {
                        "description": "createRequest",
                        "name": "createRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api_models.SessionsCreateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api_models.SessionsCreateResponse"
                        }
                    }
                }
            },
            "delete": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "session"
                ],
                "summary": "退出session。（退出登录）",
                "parameters": [
                    {
                        "description": "destroyRequest",
                        "name": "destroyRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api_models.SessionsDestroyRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api_models.SessionsDestroyResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/test/error_handler": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "test"
                ],
                "summary": "test error handler",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/api/v1/test/ping": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "test"
                ],
                "summary": "ping",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/api/v1/users/": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "获取多个用户信息，可以添加关键字对姓名搜索。",
                "parameters": [
                    {
                        "type": "integer",
                        "name": "from",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "name": "searchKeyword",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "name": "size",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api_models.UsersInfosResponse"
                        }
                    }
                }
            },
            "post": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "注册用户",
                "parameters": [
                    {
                        "description": "createRequest",
                        "name": "createRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api_models.UsersCreateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api_models.UsersCreateResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/users/{id}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "获取单个用户信息",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api_models.UsersInfoResponse"
                        }
                    }
                }
            },
            "put": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "修改用户信息",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "updateRequest",
                        "name": "updateRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api_models.UsersUpdateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api_models.UsersUpdateResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api_models.SessionsCheckResponse": {
            "type": "object",
            "properties": {
                "userID": {
                    "type": "integer"
                }
            }
        },
        "api_models.SessionsCreateRequest": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "pwd": {
                    "type": "string"
                }
            }
        },
        "api_models.SessionsCreateResponse": {
            "type": "object"
        },
        "api_models.SessionsDestroyRequest": {
            "type": "object"
        },
        "api_models.SessionsDestroyResponse": {
            "type": "object"
        },
        "api_models.User": {
            "type": "object",
            "properties": {
                "admin": {
                    "type": "boolean"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "pwd": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "api_models.UsersCreateRequest": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "pwd": {
                    "type": "string"
                }
            }
        },
        "api_models.UsersCreateResponse": {
            "type": "object"
        },
        "api_models.UsersInfoResponse": {
            "type": "object",
            "properties": {
                "admin": {
                    "type": "boolean"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "pwd": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "api_models.UsersInfosResponse": {
            "type": "object",
            "properties": {
                "infos": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/api_models.User"
                    }
                },
                "total_count": {
                    "type": "integer"
                }
            }
        },
        "api_models.UsersUpdateRequest": {
            "type": "object",
            "properties": {
                "admin": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "pwd": {
                    "type": "string"
                }
            }
        },
        "api_models.UsersUpdateResponse": {
            "type": "object"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "SoftwareWarehouse Web API",
	Description: "This is a SoftwareWarehouse API server.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}