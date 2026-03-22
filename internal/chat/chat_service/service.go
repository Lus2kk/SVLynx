package chat_service

import (
	"context"
    "fmt"
	"time"
	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/chat/chat_models"
	"github.com/svlynx/messenger/internal/chat/chat_repository"
)

type DirectService struct {
   repo chat_repository.DirectRepo
}

func NewDirectService (repo chat_repository.DirectRepo) *DirectService {
	return &DirectService{repo: repo}
}

type CreatedDirect struct {
	FirstUserID uuid.UUID `json:"first_user_id"`
	SecondUserID uuid.UUID `json:"second_user_id"`
}


func (s *DirectService) CreateNewDirectService (ctx context.Context, input CreatedDirect) (*chat_models.Direct, error) {
	// TODO: проверить существование чата
    // existing, err := s.repo.FindDirectChat(ctx, input.FirstUserId, input.SecondUserId)
    // if err == nil && existing != nil {
    //     return existing, nil
    // }
	new_direct := &chat_models.Direct{
		Id: uuid.New(),
		CreationTime: time.Now(),
	}
	new_direct.FirstMember = chat_models.ChatMember{
		ChatId: new_direct.Id,
		UserId: input.FirstUserID,
		JoinedTime: time.Now(),
	}
	new_direct.SecondMember = chat_models.ChatMember{
		ChatId: new_direct.Id,
		UserId: input.SecondUserID,
		JoinedTime: time.Now(),
	}
	 _, err := s.repo.CreateNewDirectRepo(ctx, new_direct)
    if err != nil {
        return nil, fmt.Errorf("create direct chat: %w", err)
    }
	
	return new_direct, nil
}