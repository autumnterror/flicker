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

type GenerateMDRequest struct {
	Content string `json:"content" example:"Текст документа, который нужно законспектировать"`
}

type MarkdownResponse struct {
	Markdown string `json:"markdown" example:"# Конспект ..."`
}

type N8nResponse struct {
	Output string `json:"output"`
}

// TranscribeResponse — что отдадим клиенту.
type TranscribeResponse struct {
	Text            string  `json:"text" example:"полная расшифровка аудио"`
	Filename        string  `json:"filename,omitempty" example:"lecture.mp3"`
	DurationSeconds float64 `json:"duration_seconds,omitempty" example:"123.45"`
	Language        string  `json:"language,omitempty" example:"ru"`
	Model           string  `json:"model,omitempty" example:"nova-2-general"`
}

// Вспомогательная структура для ответа Python-сервиса.
type TranscriberServiceResponse struct {
	Filename        string  `json:"filename"`
	DurationSeconds float64 `json:"duration_seconds"`
	SegmentMinutes  int     `json:"segment_minutes"`
	SegmentsCount   int     `json:"segments_count"`
	Language        string  `json:"language"`
	Model           string  `json:"model"`
	Text            string  `json:"text"`
	// Segments можно добавить при необходимости
	// Segments []struct {
	// 	SegmentIndex int    `json:"segment_index"`
	// 	StartSec     int    `json:"start_sec"`
	// 	EndSec       int    `json:"end_sec"`
	// 	Text         string `json:"text"`
	// } `json:"segments"`
}

type File2DBResponse map[string]interface{}
