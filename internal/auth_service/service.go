package auth_service

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/apperrors"
	"github.com/svlynx/messenger/internal/auth_code"
	"github.com/svlynx/messenger/internal/auth_models"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/email"
	"github.com/svlynx/messenger/internal/user_repository"
)

type Service struct {
	repo        auth_repository.AuthRepository
	userRepo    user_repository.UserRepository
	emailSender *email.Sender
}

func NewService(repo auth_repository.AuthRepository, emailSender *email.Sender, userRepo user_repository.UserRepository) *Service {
	return &Service{repo: repo, emailSender: emailSender, userRepo: userRepo}
}

func (s *Service) InitSession(ctx context.Context) (*auth_models.Session, error) {
	sessionID := uuid.New().String()
	if err := s.repo.SaveSession(ctx, sessionID); err != nil {
		slog.Error("error when creating the session", "err", err)
		return nil, err
	}
	return &auth_models.Session{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(auth_repository.SessionTTL),
	}, nil
}

func (s *Service) SendConfirmationCode(ctx context.Context, sessionID, email string) error {
	onCooldown, err := s.repo.EmailCooldownExists(ctx, email)

	if err != nil {
		return apperrors.ErrInternalError
	}

	if onCooldown {
		slog.Warn("blocked by email cooldown", "email", email)
		return apperrors.ErrEmailCooldown
	}

	attempts, err := s.repo.IncrEmailAttempts(ctx, email)

	if err != nil {
		return apperrors.ErrInternalError
	}

	if attempts > auth_repository.MaxAttempts {
		slog.Warn("number of attempts exceeded", "email", email)
		return apperrors.ErrTooManyAttempts
	}

	code, err := auth_code.GenerateSixDigitCode()

	if err != nil {
		return apperrors.ErrInternalError
	}

	err = s.repo.SavePending(ctx, sessionID, email)

	if err != nil {
		return apperrors.ErrInternalError
	}

	err = s.repo.SaveCode(ctx, email, code)

	if err != nil {
		return apperrors.ErrInternalError
	}

	if err := s.repo.SetEmailCooldown(ctx, email); err != nil {
		slog.Warn("error when setting email cooldown", "email", email)
		return apperrors.ErrInternalError
	}

	if err := s.emailSender.SendSixDigitsCode(email, code); err != nil {
		slog.Warn("Error when sending the code by email", "email", email, "err", err)
		return apperrors.ErrEmailSendFailed
	}

	slog.Info("the code has been sent by email", "email", email)
	return nil
}

func (s *Service) VerifyCode(ctx context.Context, sessionID, code string) (string, bool, error) {
	email, err := s.repo.GetPending(ctx, sessionID)

	if err != nil {
		slog.Warn("error when receiving the pending email", "sessionID", sessionID, "err", err)
		return "", false, apperrors.ErrSessionNotFound
	}

	onCooldown, err := s.repo.CodeCooldownExists(ctx, email)

	if onCooldown {
		slog.Warn("blocked by code cooldown", "email", email)
		return "", false, apperrors.ErrCodeCooldown
	}

	attempts, err := s.repo.IncrEmailAttempts(ctx, email)

	if err != nil {
		return "", false, apperrors.ErrInternalError
	}

	if attempts > auth_repository.MaxAttempts {
		if attempts > auth_repository.MaxAttempts {
			slog.Warn("number of attempts exceeded", "email", email)
			return "", false, apperrors.ErrTooManyAttempts
		}
	}

	savedCode, err := s.repo.GetCode(ctx, email)

	if err != nil {
		slog.Warn("error when getting the code from the repository", "email", email, "err", err)
		return "", false, apperrors.ErrInvalidCode
	}

	if err := s.repo.SetCodeCooldown(ctx, email); err != nil {
		slog.Warn("error when setting code cooldown", "email", email)
		return "", false, apperrors.ErrInternalError
	}

	if savedCode != code {
		slog.Warn("invalid code", "email", email)
		return "", false, apperrors.ErrInvalidCode
	}

	s.repo.DeleteCode(ctx, email)
	s.repo.DeletePending(ctx, sessionID)
	s.repo.ResetEmailAttempts(ctx, email)
	s.repo.ResetEmailAttempts(ctx, email)

	token := uuid.New().String()

	if err := s.repo.SaveAuthSession(ctx, token, email); err != nil {
		slog.Warn("error when creating a permanent session")
		return "", false, apperrors.ErrSessionCreate
	}

	exists, err := s.userRepo.UserExistsByEmail(ctx, email)

	if err != nil {
		slog.Warn("error checking the user's existence", "email", email, "err", err)
		return "", false, apperrors.ErrInternalError
	}

	slog.Info("the user has been successfully logged in", "email", email)
	return token, !exists, nil
}

func (s *Service) GetMe(ctx context.Context, token string) (*user_repository.User, error) {
	email, err := s.repo.GetAuthSession(ctx, token)

	if err != nil {
		slog.Warn("error when receiving email")
		return nil, apperrors.ErrUnauthorized
	}

	err = s.repo.RefreshAuthSession(ctx, token)
	if err != nil {
		slog.Warn("the token was not found", "err", err)
		return nil, apperrors.ErrUnauthorized
	}

	slog.Info("profile received", "email", email)

	return s.userRepo.GetUserByEmail(ctx, email)
}

func (s *Service) Logout(ctx context.Context, token string) {
	s.repo.DeleteAuthSession(ctx, token)
	slog.Info("the user goes out")
}

func (s *Service) CompleteRegistration(ctx context.Context, token, nickname, name, status string) error {
	email, err := s.repo.GetAuthSession(ctx, token)

	if err != nil {
		slog.Warn("error when receiving email to complete registration")
		return apperrors.ErrUnauthorized
	}

	exists, err := s.userRepo.NicknameExists(ctx, nickname)
	if err != nil {
		slog.Warn("error checking the existence of username", "username", nickname, "err", err)
		return apperrors.ErrInternalError
	}
	if exists {
		slog.Warn("username already exists", "username", nickname)
		return apperrors.ErrNicknameExists
	}

	colors := []string{"#2a2379", "#1D9E75", "#D85A30", "#378ADD", "#b51ed7"}
	avatar_color := colors[rand.Intn(len(colors))]

	if status == "" {
		status = "Привет!"
	}

	if err = s.userRepo.SaveUserEmail(ctx, email, nickname, name, status, avatar_color); err != nil {
		slog.Error("error when saving the profile", "email", email, "err", err)
		return err
	}

	slog.Info("profile created", "email", email, "nickname", nickname)
	return nil
}
