package auth_service

import (
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/svlynx/messenger/internal/auth_code"
	"github.com/svlynx/messenger/internal/auth_models"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/email"
)

type Service struct {
	repo        auth_repository.AuthRepository
	emailSender *email.Sender
}

func NewService(repo auth_repository.AuthRepository, emailSender *email.Sender) *Service {
	return &Service{repo: repo, emailSender: emailSender}
}

func (s *Service) InitSession() (*auth_models.Session, error) {
	sessionID := uuid.New().String()
	if err := s.repo.SaveSession(sessionID); err != nil {
		return nil, err
	}
	return &auth_models.Session{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(auth_repository.SessionTTL),
	}, nil
}

func (s *Service) SendConfirmationCode(sessionID, receiverEmail string) error {
	onCooldown, err := s.repo.CooldownExists(receiverEmail)

	if err != nil {
		return errors.New("Внутр. ошибка")
	}

	if onCooldown {
		slog.Warn("Заблокировано")
		return errors.New("Подождите 60 секунд перед повторной отправкой кода")
	}

	attempts, err := s.repo.IncrAttempts(receiverEmail)

	if err != nil {
		return errors.New("Внутр. ошибка")
	}

	if attempts > auth_repository.MaxAttempts {
		slog.Warn("Превышено кол-во попыток", "email", receiverEmail)
		return errors.New("Превышено кол-во попыток. Попробуйте позже")
	}

	code, err := auth_code.GenerateSixDigitCode()

	if err != nil {
		return errors.New("Внутр. ошибка")
	}

	err = s.repo.SavePending(sessionID, receiverEmail)

	if err != nil {
		return errors.New("Внутр. ошибка")
	}

	err = s.repo.SaveCode(receiverEmail, code)

	if err != nil {
		return errors.New("Внутр. ошибка")
	}

	s.repo.SetCooldown(receiverEmail)

	if err := s.emailSender.SendSixDigitsCode(receiverEmail, code); err != nil {
		slog.Warn("Ошибка при отправке кода на email", "email", receiverEmail, "err", err)
		return errors.New("Ошибка при отправке кода на почту")
	}

	slog.Info("Код отправлен на email", "email", receiverEmail)
	return nil
}

func (s *Service) VerifyCode(sessionID, code string) error {
	email, err := s.repo.GetPending(sessionID)

	if err != nil {
		slog.Warn("Ошибка при получении pending email", "sessionID", sessionID, "err", err)
		return errors.New("Сессия не найдена или срок действия сессии истек")
	}

	savedCode, err := s.repo.GetCode(email)

	if err != nil {
		slog.Warn("Ошибка при получении кода из репозитория", "email", email, "err", err)
		return errors.New("Неверный код или срок действия кода истек")
	}

	if savedCode != code {
		slog.Warn("Неверный код", "email", email)
		return errors.New("Неверный код")
	}

	s.repo.DeleteCode(email)
	s.repo.DeletePending(sessionID)
	s.repo.ResetAttempts(email)

	slog.Info("Пользователь успешно авторизован", "email", email)
	return nil
}
