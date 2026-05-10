package push

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, sub PushSubscription) error {
	query := `
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (endpoint) DO UPDATE
		SET user_id = $1, p256dh = $3, auth = $4, user_agent = $5, updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, sub.UserID, sub.Endpoint, sub.P256dh, sub.Auth, sub.UserAgent)
	return err
}

func (r *Repository) Delete(ctx context.Context, endpoint string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM push_subscriptions WHERE endpoint = $1`, endpoint)
	return err
}

func (r *Repository) GetByUserID(ctx context.Context, userID string) ([]PushSubscription, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, endpoint, p256dh, auth, user_agent, created_at, updated_at
		FROM push_subscriptions WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("GetByUserID: %w", err)
	}
	defer rows.Close()

	var subs []PushSubscription
	for rows.Next() {
		var s PushSubscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth, &s.UserAgent, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}
