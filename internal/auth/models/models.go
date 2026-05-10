package auth_models

import (
	"time"

	"github.com/svlynx/messenger/internal/pkg/apperrors"
)

type Session struct {
	SessionID  string    `json:"session_id" binding:"required"`
	ExpiresAt  time.Time `json:"expires_at" binding:"required"`
	Status     string    `json:"status"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	PhotoURL   string    `json:"photo_url"`
	LastName   string	 `json:"last_name"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TelegramCallbackRequest struct {
	ID        int64  `json:"id" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date" binding:"required"`
	Hash      string `json:"hash" binding:"required"`
	SessionID string `json:"session_id"`
	LastName  string `json:"last_name"`
}

func (r *TelegramCallbackRequest) Validate() error {
    if r.ID == 0 || r.FirstName == "" || r.AuthDate == 0 || r.Hash == "" {
        return apperrors.ErrInvalidRequest
    }
    return nil
}



