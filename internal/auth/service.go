package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/models"
)

type Service struct{
	repo *Repository
}

func NewService(repo *Repository) *Service{
	return &Service{repo:repo}
}

func (s *Service) InitSession() (*models.Session, error){
	sessionID := uuid.New().String()
	if err := s.repo.SaveSession(sessionID); err != nil{
		return nil,err
	}
	return &models.Session{
		SessionID : sessionID,
		ExpiresAt : time.Now().Add(SessionTTL),
	}, nil
}