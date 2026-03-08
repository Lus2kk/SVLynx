package auth

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)

const SessionTTL = 10 * time.Minute

type Repository struct {
	redis *redis.Client
}

func NewRepository(redis *redis.Client) *Repository {
	return &Repository{redis: redis}
}

func (r *Repository) SaveSession(sessionID string) error{
	ctx := context.Background()
	return r.redis.Set(
		ctx,
		"sessionID"+sessionID,
		sessionID,
		SessionTTL,
	).Err()
}
	
