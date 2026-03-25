package auth_models

import "time"

type Session struct {
	SessionID string    `json:"session_id" binding:"required"`
	ExpiresAt time.Time `json:"expires_at" binding:"required"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}