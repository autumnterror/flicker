package views

type User struct {
	Id       string `json:"id,omitempty"`
	Login    string `json:"login,omitempty"`
	Email    string `json:"email,omitempty"`
	About    string `json:"about,omitempty"`
	Photo    string `json:"photo,omitempty"`
	Password string `json:"password,omitempty"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserRegister struct {
	Login string `json:"login"`
	Email string `json:"email"`
	Pw1   string `json:"pw1"`
	Pw2   string `json:"pw2"`
}

type SWGMessage struct {
	Message string `json:"message" example:"some info"`
}

type SWGError struct {
	Error string `json:"error" example:"error"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
