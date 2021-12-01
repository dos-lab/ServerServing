package internal_models

import (
	"time"
)

type UsersCreateRequest struct {
	Name string `json:"name"`
	Pwd  string `json:"pwd"`
}

type UsersCreateResponse struct {
}

type UsersUpdateRequest struct {
	Name  string `form:"name"`
	Pwd   string `form:"pwd"`
	Admin bool   `form:"admin"`
}

type UsersUpdateResponse struct {
}

type UsersInfoResponse struct {
	*User
}

type UsersInfosRequest struct {
	SearchKeyword *string `form:"search_key_word"`
	From          int     `form:"from"`
	Size          int     `form:"size"`
}

type UsersInfosResponse struct {
	Infos      []*User `json:"infos"`
	TotalCount int     `json:"total_count"`
}

type User struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Pwd       string    `json:"pwd"`
	Admin     bool      `json:"admin"`
}
