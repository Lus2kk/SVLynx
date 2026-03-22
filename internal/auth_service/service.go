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
		slog.Error("ошибка при создании сессии", "err", err)
		return nil, err
	}
	return &auth_models.Session{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(auth_repository.SessionTTL),
	}, nil
}

func (s *Service) SendConfirmationCode(ctx context.Context, sessionID, receiverEmail string) error {
	onCooldown, err := s.repo.CooldownExists(ctx, receiverEmail)

	if err != nil {
		return apperrors.ErrInternalError
	}

	if onCooldown {
		slog.Warn("заблокировано cooldown", "email", receiverEmail)
		return apperrors.ErrCooldown
	}

	attempts, err := s.repo.IncrAttempts(ctx, receiverEmail)

	if err != nil {
		return apperrors.ErrInternalError
	}

	if attempts > auth_repository.MaxAttempts {
		slog.Warn("Превышено кол-во попыток", "email", receiverEmail)
		return apperrors.ErrTooManyAttempts
	}

	code, err := auth_code.GenerateSixDigitCode()

	if err != nil {
		return apperrors.ErrInternalError
	}

	err = s.repo.SavePending(ctx, sessionID, receiverEmail)

	if err != nil {
		return apperrors.ErrInternalError
	}

	err = s.repo.SaveCode(ctx, receiverEmail, code)

	if err != nil {
		return apperrors.ErrInternalError
	}

	s.repo.SetCooldown(ctx, receiverEmail)

	if err := s.emailSender.SendSixDigitsCode(receiverEmail, code); err != nil {
		slog.Warn("Ошибка при отправке кода на email", "email", receiverEmail, "err", err)
		return apperrors.ErrEmailSendFailed
	}

	slog.Info("Код отправлен на email", "email", receiverEmail)
	return nil
}

func (s *Service) VerifyCode(ctx context.Context, sessionID, code string) (string, bool, error) {
	email, err := s.repo.GetPending(ctx, sessionID)

	if err != nil {
		slog.Warn("Ошибка при получении pending email", "sessionID", sessionID, "err", err)
		return "", false, apperrors.ErrSessionNotFound
	}

	savedCode, err := s.repo.GetCode(ctx, email)

	if err != nil {
		slog.Warn("Ошибка при получении кода из репозитория", "email", email, "err", err)
		return "", false, apperrors.ErrInvalidCode
	}

	if savedCode != code {
		slog.Warn("Неверный код", "email", email)
		return "", false, apperrors.ErrInvalidCode
	}

	s.repo.DeleteCode(ctx, email)
	s.repo.DeletePending(ctx, sessionID)
	s.repo.ResetAttempts(ctx, email)

	token := uuid.New().String()

	if err := s.repo.SaveAuthSession(ctx, token, email); err != nil {
		slog.Warn("ошибка при создании постоянной сессии")
		return "", false, apperrors.ErrSessionCreate
	}

	exists, err := s.userRepo.UserExistsByEmail(ctx, email)

	if err != nil {
		slog.Warn("Ошибка при проверке существования пользователя", "email", email, "err", err)
		return "", false, apperrors.ErrInternalError
	}

	slog.Info("Пользователь успешно авторизован", "email", email)
	return token, !exists, nil
}

func (s *Service) GetMe(ctx context.Context, token string) (*user_repository.User, error) {
	email, err := s.repo.GetAuthSession(ctx, token)

	if err != nil {
		slog.Warn("Ошибка при получении email")
		return nil, apperrors.ErrUnauthorized
	}

	err = s.repo.RefreshAuthSession(ctx, token)
	if err != nil{
		slog.Warn("токен не найден", "err", err)
		return nil, apperrors.ErrUnauthorized
	}

	slog.Info("получен профиль", "email", email)

	return s.userRepo.GetUserByEmail(ctx, email)
}

func (s *Service) Logout(ctx context.Context, token string) {
	s.repo.DeleteAuthSession(ctx, token)
	slog.Info("пользователь вышел")
}

func (s *Service) CompleteRegistration(ctx context.Context, token, nickname, name, status string) error {
	email, err := s.repo.GetAuthSession(ctx, token)

	if err != nil {
		slog.Warn("Ошибка при получении email для завершения регистрации")
		return apperrors.ErrUnauthorized
	}

	exists, err := s.userRepo.NicknameExists(ctx, nickname)
	if err != nil {
		slog.Warn("Ошибка при проверке существования username", "username", nickname, "err", err)
		return apperrors.ErrInternalError
	}
	if exists {
		slog.Warn("Username уже существует", "username", nickname)
		return apperrors.ErrNicknameExists
	}

	colors := []string{"#7F77DD", "#1D9E75", "#D85A30", "#378ADD", "#b51ed7"}
	avatar_color := colors[rand.Intn(len(colors))]

	if status == "" {
		status = "Привет!"
	}

	if err = s.userRepo.SaveUserEmail(ctx, email, nickname, name, status, avatar_color); err != nil {
    	slog.Error("ошибка при сохранении профиля", "email", email, "err", err)
    	return err
	}
	
	slog.Info("профиль создан", "email", email, "nickname", nickname)
	return nil
}
