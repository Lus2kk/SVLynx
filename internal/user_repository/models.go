package user_repository

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	TelegramID  int64  `json:"telegram_id"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhotoURL    string `json:"photo_url"`
	Nickname    string `json:"nickname"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	AvatarColor string `json:"avatar_color"`
}
