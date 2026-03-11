package auth_service

import (
	"time"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/auth_models"
	"github.com/svlynx/messenger/internal/auth_repository"
)

type Service struct{
	repo auth_repository.AuthRepository
}

func NewService(repo auth_repository.AuthRepository) *Service{
	return &Service{repo:repo}
}

func (s *Service) InitSession() (*auth_models.Session, error){
	sessionID := uuid.New().String()
	if err := s.repo.SaveSession(sessionID); err != nil{
		return nil,err
	}
	return &auth_models.Session{
		SessionID : sessionID,
		ExpiresAt : time.Now().Add(auth_repository.SessionTTL),
	}, nil
}