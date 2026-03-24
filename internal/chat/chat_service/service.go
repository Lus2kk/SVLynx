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

func NewDirectService(repo chat_repository.DirectRepo) *DirectService {
	return &DirectService{repo: repo}
}

type CreatedDirect struct {
	FirstUserID  uuid.UUID `json:"first_user_id"`
	SecondUserID uuid.UUID `json:"second_user_id"`
}

type MessageService struct {
	repo chat_repository.MessageRepo
}

func NewMessageService(repo chat_repository.MessageRepo) *MessageService {
	return &MessageService{repo: repo}
}

type CreatedMessage struct {
	ChatID   uuid.UUID `json:"chat_id"`
	SenderID uuid.UUID `json:"sender_id"`
	Content  string    `json:"content"`
}

func (s *DirectService) CreateNewDirectService(ctx context.Context, input CreatedDirect) (*chat_models.Direct, error) {
	existing, err := s.repo.GetDirectByIdRepo(ctx, input.FirstUserID, input.SecondUserID)
	if err == nil && existing != nil {
		return existing, nil
	}
	new_direct := &chat_models.Direct{
		Id:           uuid.New(),
		CreationTime: time.Now(),
	}
	new_direct.FirstMember = chat_models.ChatMember{
		ChatId:     new_direct.Id,
		UserId:     input.FirstUserID,
		JoinedTime: time.Now(),
	}
	new_direct.SecondMember = chat_models.ChatMember{
		ChatId:     new_direct.Id,
		UserId:     input.SecondUserID,
		JoinedTime: time.Now(),
	}
	_, err = s.repo.CreateNewDirectRepo(ctx, new_direct)
	if err != nil {
		return nil, fmt.Errorf("create direct chat: %w", err)
	}

	return new_direct, nil
}

func (s *DirectService) GetDirectById(ctx context.Context, MyId uuid.UUID, CompanionId uuid.UUID) (*chat_models.Direct, error) {
	direct, err := s.repo.GetDirectByIdRepo(ctx, MyId, CompanionId)
	if err != nil {
		return nil, fmt.Errorf("troubles with finding: %w", err)
	}
	if direct == nil {
		return nil, fmt.Errorf("chat not found")
	}
	return direct, nil
}

func (s *DirectService) GetListOfDirectsByIDService(ctx context.Context, user_id uuid.UUID) ([]*chat_models.Direct, error) {
	directs, err := s.repo.GetListOfDirectsListByIDRepo(ctx, user_id)
	if err != nil {
		return nil, fmt.Errorf("troubles with findong: %w", err)
	}
	return directs, nil
}

func (s *MessageService) SendMessage(ctx context.Context, input CreatedMessage) (*chat_models.Message, error) {
	message := &chat_models.Message{
		ID:        uuid.New(),
		ChatID:    input.ChatID,
		SenderID:  input.SenderID,
		Content:   input.Content,
		Status:    chat_models.Sent,
		CreatedAT: time.Now(),
	}

	result, err := s.repo.SendMessageRepo(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("send message error: %w", err)
	}

	return result, nil
}

func (s *MessageService) GetMessagesByChatIdService(ctx context.Context, chatId uuid.UUID, before time.Time, limit int) ([]*chat_models.Message, error) {
	messages, err := s.repo.GetMessagesByChatIdRepo(ctx, chatId, before, limit)
	if err != nil {
		return nil, fmt.Errorf("get messages error: %w", err)
	}
	return messages, nil
}
