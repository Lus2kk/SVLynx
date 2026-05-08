package chat_service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	chat_models "github.com/svlynx/messenger/internal/chat/models"
	chat_repository "github.com/svlynx/messenger/internal/chat/repository"
	user_models "github.com/svlynx/messenger/internal/user/models"
	user_repository "github.com/svlynx/messenger/internal/user/repository"
)

type DirectService struct {
	repo      chat_repository.DirectRepo
	user_repo user_repository.UserRepository
}

func NewDirectService(repo chat_repository.DirectRepo, user_repo user_repository.UserRepository) *DirectService {
	return &DirectService{
		repo:      repo,
		user_repo: user_repo,
	}
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
	ChatID      uuid.UUID               `json:"chat_id"`
	SenderID    uuid.UUID               `json:"sender_id"`
	RecipientID uuid.UUID               `json:"recipient_id"`
	Content     string                  `json:"content"`
	Type        chat_models.MessageType `json:"type"`
	Duration    int                     `json:"duration"`
}

func (s *DirectService) CreateNewDirectService(ctx context.Context, input CreatedDirect) (*chat_models.Direct, error) {
	if input.FirstUserID == input.SecondUserID {
		return nil, fmt.Errorf("cannot create direct chat with yourself")
	}

	existing, err := s.repo.GetDirectByIdRepo(ctx, input.FirstUserID, input.SecondUserID)
	if err != nil {
		return nil, fmt.Errorf("check existing direct error: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	now := time.Now()

	newDirect := &chat_models.Direct{
		Id:           uuid.New(),
		CreationTime: now,
		FirstMember: chat_models.ChatMember{
			ChatId:     uuid.Nil,
			UserId:     input.FirstUserID,
			JoinedTime: now,
		},
		SecondMember: chat_models.ChatMember{
			ChatId:     uuid.Nil,
			UserId:     input.SecondUserID,
			JoinedTime: now,
		},
	}

	newDirect.FirstMember.ChatId = newDirect.Id
	newDirect.SecondMember.ChatId = newDirect.Id

	result, err := s.repo.CreateNewDirectRepo(ctx, newDirect)
	if err != nil {
		return nil, fmt.Errorf("create direct error: %w", err)
	}

	return result, nil
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

func (s *DirectService) GetListOfDirectsByIDService(ctx context.Context, userId uuid.UUID) ([]*chat_models.DirectListItem, error) {
	directs, err := s.repo.GetListOfDirectsListByIDRepo(ctx, userId)
	if err != nil {
		return nil, err
	}
	return directs, nil
}

func (s *DirectService) DeleteDirectService(ctx context.Context, chatID uuid.UUID) error {
	if err := s.repo.DeleteDirectRepo(ctx, chatID); err != nil {
		return fmt.Errorf("delete direct error: %w", err)
	}
	return nil
}

func (s *MessageService) SendMessage(ctx context.Context, input CreatedMessage) (*chat_models.Message, error) {
	msgType := input.Type
    if msgType == "" {
        msgType = chat_models.TextMessage
    }
	
	message := &chat_models.Message{
		ID:        uuid.New(),
		ChatID:    input.ChatID,
		SenderID:  input.SenderID,
		Content:   input.Content,
		Status:    chat_models.Sent,
		CreatedAT: time.Now(),
		Type: msgType,
		Duration: input.Duration,
	}

	result, err := s.repo.SendMessageRepo(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("send message error: %w", err)
	}

	return result, nil
}

func (s *MessageService) GetMessagesByChatIdService(ctx context.Context, chat_Id uuid.UUID, before time.Time, limit int) ([]*chat_models.Message, error) {
	messages, err := s.repo.GetMessagesByChatIdRepo(ctx, chat_Id, before, limit)
	if err != nil {
		return nil, fmt.Errorf("get messages error: %w", err)
	}
	return messages, nil
}

func (s *MessageService) SearchMesaageService(ctx context.Context, chat_id uuid.UUID, content string) ([]*chat_models.Message, error) {
	if content == "" {
		return nil, fmt.Errorf("string cannot be empty")
	}
	messages, err := s.repo.SearchMesageRepo(ctx, chat_id, content)
	if err != nil {
		return nil, fmt.Errorf("get message error: %w", err)
	}
	return messages, nil
}

func (s *MessageService) UpdateMessageStatusService(ctx context.Context, status chat_models.MessageStatus, message_id uuid.UUID) error {
	if err := s.repo.UpdateMessageStatusRepo(ctx, message_id, status); err != nil {
		return fmt.Errorf("error of updating message status : %w", err)
	}
	return nil
}

func (s *MessageService) DeleteMessageService(ctx context.Context, message_id uuid.UUID) error {
	if err := s.repo.DeleteMessageRepo(ctx, message_id); err != nil {
		return fmt.Errorf("delete message error : %w", err)
	}
	return nil
}

func (s *DirectService) SearchUsersService(ctx context.Context, currentUserID string, query string) ([]*user_models.User, error) {
	if query == "" {
		return []*user_models.User{}, nil
	}

	limit := 20
	return s.user_repo.SearchUsers(ctx, currentUserID, query, limit)
}

func (s *MessageService) MarkChatMessagesAsReadService(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error {
	if err := s.repo.MarkChatMessagesAsReadRepo(ctx, chatID, userID); err != nil {
		return fmt.Errorf("mark messages as read service error: %w", err)
	}
	return nil
}

func (s *DirectService) GetUserStatusService(ctx context.Context, userID uuid.UUID) (bool, time.Time, error) {
	isOnline, lastSeen, err := s.user_repo.GetUserStatus(ctx, userID.String())
	if err != nil {
		return false, time.Time{}, fmt.Errorf("get user status error: %w", err)
	}
	return isOnline, lastSeen, nil
}

func (s *DirectService) UpdateLastSeenService(ctx context.Context, userID uuid.UUID) {
	_ = s.user_repo.SetUserOnlineStatus(ctx, userID.String(), true)
}

func (s *DirectService) SetUserOnline(ctx context.Context, userID uuid.UUID) error {
	return s.user_repo.SetUserOnlineStatus(ctx, userID.String(), true)
}

func (s *DirectService) SetUserOffline(ctx context.Context, userID uuid.UUID) error {
	return s.user_repo.SetUserOnlineStatus(ctx, userID.String(), false)
}
