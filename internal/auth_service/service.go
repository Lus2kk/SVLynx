package auth_service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/apperrors"
	"github.com/svlynx/messenger/internal/auth_models"
	"github.com/svlynx/messenger/internal/auth_repository"
)
type AuthRepository interface {
	SaveSession(ctx context.Context, sessionID string) error
	UpdateSession(ctx context.Context, session *auth_models.Session) error
	GetSession(ctx context.Context, sessionID string) (*auth_models.Session, error)
} 

type Service struct {
	repo          AuthRepository
}

func NewService(repo AuthRepository) *Service {
	return &Service{
		repo:          repo,
	}
}

func (s *Service) InitSession(ctx context.Context) (*auth_models.Session, error) {
	sessionID := uuid.New().String()
	if err := s.repo.SaveSession(ctx, sessionID); err != nil {
		return nil, err
	}
	return &auth_models.Session{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(auth_repository.SessionTTL),
		Status:    auth_models.StatusPending,
	}, nil
}

func (s *Service) TelegramCallback(ctx context.Context, telegramToken string, req *auth_models.TelegramCallbackRequest) (*auth_models.Session, error) {
	if !s.verifyHash(telegramToken, req) {
		return nil, apperrors.ErrInvalidHash

	}

	if time.Now().Unix()-req.AuthDate > 86400 {
		return nil, apperrors.ErrAuthExpired
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
		PhotoURL:   req.PhotoURL,
		LastName: req.LastName,
	}

	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// TODO: refactor verifyHash to be more concise
func (s *Service) verifyHash(telegramToken string, req *auth_models.TelegramCallbackRequest) bool {
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
	if req.LastName != ""{
		data["last_name"] = req.LastName
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
	h.Write([]byte(telegramToken))
	secretKey := h.Sum(nil)
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	return expectedHash == req.Hash
}