package chat_repository

import (
	"context"

	"time"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/chat/chat_models"
)

type DirectRepo interface {
	CreateNewDirectRepo(ctx context.Context, chat *chat_models.Direct) (*chat_models.Direct, error)
	GetDirectByIdRepo(ctx context.Context, MYid uuid.UUID, CompanionId uuid.UUID) (*chat_models.Direct, error)
	GetListOfDirectsListByIDRepo(ctx context.Context, user_id uuid.UUID) ([]*chat_models.Direct, error)
}

type MessageRepo interface {
	SendMessageRepo(ctx context.Context, message *chat_models.Message) (*chat_models.Message, error)
	GetMessagesByChatIdRepo(ctx context.Context, chatId uuid.UUID, before time.Time, limit int) ([]*chat_models.Message, error)
	SearchMesageRepo(ctx context.Context, chat_id uuid.UUID, content string) ([]*chat_models.Message, error)
	UpdateMessageStatusRepo(ctx context.Context, message_id uuid.UUID, status chat_models.MessageStatus) error 
}
