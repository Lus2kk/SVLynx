package chat_repository

import (
	"context"

	"github.com/svlynx/messenger/internal/chat/chat_models"
)

type DirectRepo interface {
	CreateNewDirectRepo(ctx context.Context, chat *chat_models.Direct) (*chat_models.Direct, error)
}
