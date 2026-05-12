package user_models

import "time"

type User struct {
    ID               string    `json:"id"`
    Email            string    `json:"email"`
    TelegramID       int64     `json:"telegram_id"`
    Username         string    `json:"username"`
    FirstName        string    `json:"first_name"`
    LastName         string    `json:"last_name"`
    PhotoURL         string    `json:"photo_url"`
    Nickname         string    `json:"nickname"`
    Name             string    `json:"name"`
    Status           string    `json:"status"`
    AvatarColor      string    `json:"avatar_color"`
    ProfileCompleted bool      `json:"profile_completed"`
    LastSeen         time.Time `json:"last_seen"`
    IsOnline         bool      `json:"is_online"`
    IsDeveloper      bool      `json:"is_developer"`
}
