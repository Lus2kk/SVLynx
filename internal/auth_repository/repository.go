package auth_repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	SessionTTL = 10 * time.Minute

	// email consts
	CodeTTL     = 3 * time.Minute
	PendingTTL  = 10 * time.Minute
	CooldownTTL = 1 * time.Minute
	AttemptsTTL = 15 * time.Minute
	MaxAttempts = 5
)

type AuthRepository interface {
	SaveSession(sessionID string) error

	SaveCode(email string, code string) error
	GetCode(email string) (string, error)
	DeleteCode(email string) error

	SavePending(sessionID, email string) error
	GetPending(sessionID string) (string, error)
	DeletePending(sessionID string) error

	CooldownExists(email string) (bool, error)
	SetCooldown(email string) error

	IncrAttempts(email string) (int64, error)
	ResetAttempts(email string) error
}

type Repository struct {
	redis *redis.Client
}

func NewRepository(redis *redis.Client) *Repository {
	return &Repository{redis: redis}
}

func (r *Repository) SaveSession(sessionID string) error {
	ctx := context.Background()
	return r.redis.Set(
		ctx,
		"session:"+sessionID,
		sessionID,
		SessionTTL,
	).Err()
}

func (r *Repository) SaveCode(email string, code string) error {
	ctx := context.Background()

	return r.redis.Set(ctx, "code:"+email, code, CodeTTL).Err()
}

func (r *Repository) GetCode(email string) (string, error) {
	ctx := context.Background()

	return r.redis.Get(ctx, "code:"+email).Result()
}

func (r *Repository) DeleteCode(email string) error {
	ctx := context.Background()

	return r.redis.Del(ctx, "code:"+email).Err()
}

func (r *Repository) SavePending(sessionID, email string) error {
	ctx := context.Background()

	return r.redis.Set(ctx, "pending:"+sessionID, email, PendingTTL).Err()
}

func (r *Repository) GetPending(sessionID string) (string, error) {
	ctx := context.Background()

	return r.redis.Get(ctx, "pending:"+sessionID).Result()
}

func (r *Repository) DeletePending(sessionID string) error {
	ctx := context.Background()

	return r.redis.Del(ctx, "pending:"+sessionID).Err()
}

func (r *Repository) CooldownExists(email string) (bool, error) {
	ctx := context.Background()

	n, err := r.redis.Exists(ctx, "cooldown"+email).Result()
	return n > 0, err
}

func (r *Repository) SetCooldown(email string) error {
	ctx := context.Background()
	return r.redis.Set(ctx, "cooldown:"+email, 1, CooldownTTL).Err()
}

func (r *Repository) IncrAttempts(email string) (int64, error) {
	ctx := context.Background()

	count, err := r.redis.Incr(ctx, "attempts:"+email).Result()

	if err != nil {
		return 0, err
	}

	if count == 1 {
		r.redis.Expire(ctx, "attempts:"+email, AttemptsTTL)
	}

	return count, nil
}

func (r *Repository) ResetAttempts(email string) error {
	ctx := context.Background()

	return r.redis.Del(ctx, "attempts:"+email).Err()
}