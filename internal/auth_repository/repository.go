package auth_repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth_models"
)

const SessionTTL = 10 * time.Minute


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

func (r *Repository) UpdateSession(ctx context.Context, session *auth_models.Session) error {

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return r.redis.Set(
		ctx,
		"session:"+session.SessionID,
		data,
		SessionTTL,
	).Err()
}
func (r *Repository) GetSession(ctx context.Context, sessionID string) (*auth_models.Session, error) {

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
