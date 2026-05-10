package push

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type Sender struct {
	repo       *Repository
	privateKey string
	publicKey  string
	email      string
}

func NewSender(repo *Repository, privateKey, publicKey, email string) *Sender {
	return &Sender{
		repo:       repo,
		privateKey: privateKey,
		publicKey:  publicKey,
		email:      email,
	}
}

type PushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
}

func (s *Sender) SendToUser(ctx context.Context, userID string, payload PushPayload) error {
	subs, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("SendToUser: get subscriptions: %w", err)
	}

	if len(subs) == 0 {
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("SendToUser: marshal payload: %w", err)
	}

	for _, sub := range subs {
		webSub := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}

		resp, err := webpush.SendNotification(data, webSub, &webpush.Options{
			VAPIDPublicKey:  s.publicKey,
			VAPIDPrivateKey: s.privateKey,
			Subscriber:      s.email,
			TTL:             30,
		})

		if err != nil {
			slog.Warn("SendPush: failed to send", "endpoint", sub.Endpoint, "err", err)
			if resp != nil && (resp.StatusCode == 404 || resp.StatusCode == 410) {
				_ = s.repo.Delete(ctx, sub.Endpoint)
			}
			continue
		}
		resp.Body.Close()
		slog.Info("SendPush: sent", "userID", userID, "status", resp.StatusCode)
	}

	return nil
}