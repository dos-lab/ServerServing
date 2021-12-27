package internal_models

type SessionsCreateRequest struct {
	Name string
	Pwd  string
}

type SessionsCreateResponse struct {
	Token string `json:"token"`
}

type SessionsDestroyRequest struct {
}

type SessionsDestroyResponse struct {
}

type SessionsCheckRequest struct {
}

type SessionsCheckResponse struct {
	UserID int
}
