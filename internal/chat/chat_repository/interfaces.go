package chat_repository

import (
	"context"

	"github.com/svlynx/messenger/internal/chat/chat_models"
)


type ChatRepo interface{
CreateNewChat (ctx context.Context, chat *chat_models.Chat) (*chat_models.Chat, error) //пока нет бд не могу записать 
}
