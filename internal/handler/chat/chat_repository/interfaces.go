package chat_repository

import (
	"context"

	"github.com/svlynx/messenger/internal/handler/chat/chat_models"
)


type ChatRepo interface{
CreateNewChat (ctx context.Context, chat *chat_models.Chat) error //пока нет бд не могу записать 
}

type MemberRepo interface {
AddMember (ctx context.Context, member *chat_models.ChatMember) error // пока нет бд не могу записать 
}