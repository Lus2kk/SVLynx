package chat_service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/handler/chat/chat_models"
	"github.com/svlynx/messenger/internal/handler/chat/chat_repository"
)

type ChatService struct {
	ChatRepo   chat_repository.ChatRepo
	MemberRepo chat_repository.MemberRepo
}

type CreatedChat struct {
	Name *string                     `json:"name"`
	Type chat_models.ChatType        `json:"type"`
	OwnerID uuid.UUID                `json:"owner_id"`
}


func (service *ChatService) CreateNewChat (ctx context.Context, chat CreatedChat) (*chat_models.Chat, error){
    if len(*chat.Name) < 1 && chat.Name == nil {
          return nil, errors.New("Incorrect name of the chat!")
	}
	
	
   switch chat.Type {
    case chat_models.Direct, chat_models.Channel , chat_models.Group:
    default:
        return nil, errors.New("invalid chat type")
    }
	createdChat := chat_models.Chat{
		Id: uuid.New(),
		Type: chat.Type,
		Name: *chat.Name,
		OwnerID: chat.OwnerID,
	}
	if err := service.ChatRepo.CreateNewChat(ctx, &createdChat); err != nil {
  return nil, errors.New("chat not saved in database!")
	}
	return &createdChat, nil 
	 
}
