package auth_repository

import (
	"context"
	"log/slog"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth_models"
)

const (
	SessionTTL = 10 * time.Minute
	CodeTTL          = 3 * time.Minute
	PendingTTL       = 10 * time.Minute
	EmailCooldownTTL = 1 * time.Minute
	CodeCooldownTTL  = 5 * time.Second
	AttemptsTTL      = 15 * time.Minute
	RefreshTokenTTL  = 30 * 24 * time.Hour
	AccesTokenTTL	 = 15 * time.Minute
	MaxAttempts      = 5
)

type AuthRepository interface {
	SaveSession(ctx context.Context, sessionID string) error
	GetSessionEmail(ctx context.Context, sessionID string) (string, error)
	GetSessionTG(ctx context.Context, sesionID string) (*auth_models.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	UpdateSession(ctx context.Context, session *auth_models.Session) error

	SaveCode(ctx context.Context, email string, code string) error
	GetCode(ctx context.Context, email string) (string, error)
	DeleteCode(ctx context.Context, email string) error

	SavePending(ctx context.Context, sessionID, email string) error
	GetPending(ctx context.Context, sessionID string) (string, error)
	DeletePending(ctx context.Context, sessionID string) error

	EmailCooldownExists(ctx context.Context, email string) (bool, error)
	SetEmailCooldown(ctx context.Context, email string) error

	CodeCooldownExists(ctx context.Context, email string) (bool, error)
	SetCodeCooldown(ctx context.Context, email string) error

	IncrEmailAttempts(ctx context.Context, email string) (int64, error)
	ResetEmailAttempts(ctx context.Context, email string) error

	InrcCodeAttempts(ctx context.Context, email string) (int64, error)
	ResetCodeAttempts(ctx context.Context, email string) error

	SaveRefreshToken(ctx context.Context, token, email string) error
	GetRefreshToken(ctx context.Context, token string) (string, error) 
	DeleteRefreshToken(ctx context.Context, token string) error
}

type Repository struct {
	redis *redis.Client
}

func NewRepository(redis *redis.Client) *Repository {
	return &Repository{redis: redis}
}

func (r *Repository) SaveSession(ctx context.Context, sessionID string) error {
	return r.redis.Set(ctx, "session_id:"+sessionID, sessionID, SessionTTL).Err()
}

func (r *Repository) GetSessionEmail(ctx context.Context, sessionID string) (string, error) {
	return r.redis.Get(ctx, "session_id:"+sessionID).Result()
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.redis.Del(ctx, "session_id:"+sessionID).Err()
}

func (r *Repository) SaveCode(ctx context.Context, email string, code string) error {
	return r.redis.Set(ctx, "code:"+email, code, CodeTTL).Err()
}

func (r *Repository) GetCode(ctx context.Context, email string) (string, error) {
	return r.redis.Get(ctx, "code:"+email).Result()
}

func (r *Repository) DeleteCode(ctx context.Context, email string) error {
	return r.redis.Del(ctx, "code:"+email).Err()
}

func (r *Repository) SavePending(ctx context.Context, sessionID, email string) error {
	return r.redis.Set(ctx, "pending:"+sessionID, email, PendingTTL).Err()
}

func (r *Repository) GetPending(ctx context.Context, sessionID string) (string, error) {
	return r.redis.Get(ctx, "pending:"+sessionID).Result()
}

func (r *Repository) DeletePending(ctx context.Context, sessionID string) error {
	return r.redis.Del(ctx, "pending:"+sessionID).Err()
}

func (r *Repository) EmailCooldownExists(ctx context.Context, email string) (bool, error) {
	n, err := r.redis.Exists(ctx, "email_cooldown:"+email).Result()
	return n > 0, err
}

func (r *Repository) SetEmailCooldown(ctx context.Context, email string) error {
	return r.redis.Set(ctx, "email_cooldown:"+email, 1, EmailCooldownTTL).Err()
}

func (r *Repository) CodeCooldownExists(ctx context.Context, email string) (bool, error) {
	n, err := r.redis.Exists(ctx, "code_cooldown:"+email).Result()
	return n > 0, err
}

func (r *Repository) SetCodeCooldown(ctx context.Context, email string) error {
	return r.redis.Set(ctx, "code_cooldown:"+email, 1, CodeCooldownTTL).Err()
}

func (r *Repository) IncrEmailAttempts(ctx context.Context, email string) (int64, error) {
	count, err := r.redis.Incr(ctx, "email_attempts:"+email).Result()

	if err != nil {
		return 0, err
	}

	if count == 1 {
		if err := r.redis.Expire(ctx, "email_attempts:"+email, AttemptsTTL); err != nil {
			slog.Warn("couldn't set TTL for email attempts", "email", email, "err", err)
		}
	}

	return count, nil
}

func (r *Repository) ResetEmailAttempts(ctx context.Context, email string) error {
	return r.redis.Del(ctx, "email_attempts:"+email).Err()
}

func (r *Repository) InrcCodeAttempts(ctx context.Context, email string) (int64, error) {
	count, err := r.redis.Incr(ctx, "code_attempts:"+email).Result()


	if err != nil {
		return 0, err
	}

	if count == 1 {
		if err := r.redis.Expire(ctx, "code_attempts:"+email, AttemptsTTL); err != nil {
			slog.Warn("couldn't set TTl for code attempts", "email", email)
		}
	}

	return count, nil
}

func (r *Repository) ResetCodeAttempts(ctx context.Context, email string) error {
	return r.redis.Del(ctx, "code_attempts:"+email).Err()
}

func (r *Repository) SaveRefreshToken(ctx context.Context, token, email string) error {
	return r.redis.Set(ctx, "refresh:"+token, email, RefreshTokenTTL).Err()
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (string, error){
	return r.redis.Get(ctx, "refresh:"+token).Result()
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.redis.Del(ctx, "refresh:"+token).Err()
}

func (r *Repository) UpdateSession(ctx context.Context, session *auth_models.Session) error {

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, "session:"+session.SessionID, data, SessionTTL).Err()
}

func (r *Repository) GetSessionTG(ctx context.Context, sessionID string) (*auth_models.Session, error) {

	data, err := r.redis.Get(ctx, "session:"+sessionID).Bytes()

	if err != nil {
		return nil, err
	}

	var session auth_models.Session
	if err := json.Unmarshal(data, &session); err != nil{
		return nil, err
	}
	return &session, nil

}
