package auth_repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	SessionTTL = 10 * time.Minute

	CodeTTL     = 3 * time.Minute
	PendingTTL  = 10 * time.Minute
	CooldownTTL = 1 * time.Minute
	AttemptsTTL = 15 * time.Minute
	AuthSessionTTL = 30 * 24 * time.Hour
	MaxAttempts = 5
)

type AuthRepository interface {
	SaveSession(ctx context.Context, sessionID string) error

	SaveCode(ctx context.Context, email string, code string) error
	GetCode(ctx context.Context, email string) (string, error)
	DeleteCode(ctx context.Context, email string) error

	SavePending(ctx context.Context, sessionID, email string) error
	GetPending(ctx context.Context, sessionID string) (string, error)
	DeletePending(ctx context.Context, sessionID string) error

	CooldownExists(ctx context.Context, email string) (bool, error)
	SetCooldown(ctx context.Context, email string) error

	IncrAttempts(ctx context.Context, email string) (int64, error)
	ResetAttempts(ctx context.Context, email string) error

	SaveAuthSession(ctx context.Context, token, email string) error
	GetAuthSession(ctx context.Context, token string) (string, error)
	DeleteAuthSession(ctx context.Context, token string) error
	RefreshAuthSession(ctx context.Context, token string) error
}

type Repository struct {
	redis *redis.Client
}

func NewRepository(redis *redis.Client) *Repository {
	return &Repository{redis: redis}
}

func (r *Repository) SaveSession(ctx context.Context, sessionID string) error {
	return r.redis.Set(
		ctx,
		"session:"+sessionID,
		sessionID,
		SessionTTL,
	).Err()
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

func (r *Repository) CooldownExists(ctx context.Context, email string) (bool, error) {
	n, err := r.redis.Exists(ctx, "cooldown:"+email).Result()
	return n > 0, err
}

func (r *Repository) SetCooldown(ctx context.Context, email string) error {
	return r.redis.Set(ctx, "cooldown:"+email, 1, CooldownTTL).Err()
}

func (r *Repository) IncrAttempts(ctx context.Context, email string) (int64, error) {
	count, err := r.redis.Incr(ctx, "attempts:"+email).Result()

	if err != nil {
		return 0, err
	}

	if count == 1 {
		if err := r.redis.Expire(ctx, "attempts:"+email, AttemptsTTL); err != nil{
			slog.Warn("не удалось выставить TTL для attempts", "email", email, "err", err)
		}
	}

	return count, nil
}

func (r *Repository) ResetAttempts(ctx context.Context, email string) error {
	return r.redis.Del(ctx, "attempts:"+email).Err()
}

func (r *Repository) SaveAuthSession(ctx context.Context, token, email string) error {
	return r.redis.Set(ctx, "auth:" + token, email, AuthSessionTTL).Err()
}

func (r *Repository) GetAuthSession(ctx context.Context, token string) (string, error) {
	return r.redis.Get(ctx, "auth:" + token).Result()
}

func (r *Repository) DeleteAuthSession(ctx context.Context, token string) error {
	return r.redis.Del(ctx, "auth:" + token).Err()
}

func (r *Repository) RefreshAuthSession (ctx context.Context, token string) error {
	return r.redis.Expire(ctx, "auth:" + token, AuthSessionTTL).Err()
}