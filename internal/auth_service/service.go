package auth_service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/auth_models"
	"github.com/svlynx/messenger/internal/auth_repository"
)

type Service struct {
	repo          auth_repository.AuthRepository
	telegramToken string
}

func NewService(repo auth_repository.AuthRepository, telegramToken string) *Service {
	return &Service{
		repo:          repo,
		telegramToken: telegramToken,
	}
}

func (s *Service) InitSession() (*auth_models.Session, error) {
	sessionID := uuid.New().String()
	if err := s.repo.SaveSession(sessionID); err != nil {
		return nil, err
	}
	return &auth_models.Session{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(auth_repository.SessionTTL),
		Status:    auth_models.StatusPending,
	}, nil
}

func (s *Service) TelegramCallback(req *auth_models.TelegramCallbackRequest) (*auth_models.Session, error) {
	if !s.verifyHash(req) {
		return nil, errors.New("invalid hash")
	}

	if time.Now().Unix()-req.AuthDate > 86400 {
		return nil, errors.New("auth data expired")
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	session := &auth_models.Session{
		SessionID:  sessionID,
		ExpiresAt:  time.Now().Add(auth_repository.SessionTTL),
		Status:     auth_models.StatusApproved,
		TelegramID: req.ID,
		Username:   req.Username,
		FirstName:  req.FirstName,
	}

	if err := s.repo.UpdateSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) verifyHash(req *auth_models.TelegramCallbackRequest) bool {
	data := map[string]string{
		"id":         strconv.FormatInt(req.ID, 10),
		"first_name": req.FirstName,
		"auth_date":  strconv.FormatInt(req.AuthDate, 10),
	}
	if req.Username != "" {
		data["username"] = req.Username
	}
	if req.PhotoURL != "" {
		data["photo_url"] = req.PhotoURL
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+data[k])
	}
	dataString := strings.Join(parts, "\n")

	h := sha256.New()
	h.Write([]byte(s.telegramToken))
	secretKey := h.Sum(nil)
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return expectedHash == req.Hash
}