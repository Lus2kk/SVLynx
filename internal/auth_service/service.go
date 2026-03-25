package auth_service

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/apperrors"
	"github.com/svlynx/messenger/internal/auth_code"
	"github.com/svlynx/messenger/internal/auth_jwt"
	"github.com/svlynx/messenger/internal/auth_models"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/email"
	"github.com/svlynx/messenger/internal/user_repository"
)

type Service struct {
	repo        auth_repository.AuthRepository
	userRepo    user_repository.UserRepository
	emailSender *email.Sender
	jwtSecret	string
}

func NewService(repo auth_repository.AuthRepository, emailSender *email.Sender, userRepo user_repository.UserRepository, jwtSecret string) *Service {
	return &Service{
		repo: repo,
		emailSender: emailSender,
		userRepo: userRepo,
		jwtSecret: jwtSecret,
	}
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
	_, err := s.repo.GetSession(ctx, sessionID)
    if err != nil {
        if errors.Is(err, redis.Nil) {
            return apperrors.ErrSessionNotFound
        }
        return apperrors.ErrSessionCreate
    }

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

func (s *Service) VerifyCode(ctx context.Context, sessionID, code string) (*auth_models.TokenPair, bool, error) {
	email, err := s.repo.GetPending(ctx, sessionID)

	if err != nil {
		slog.Warn("error when receiving the pending email", "sessionID", sessionID, "err", err)
		return nil, false, apperrors.ErrSessionNotFound
	}

	onCooldown, err := s.repo.CodeCooldownExists(ctx, email)

	if onCooldown {
		slog.Warn("blocked by code cooldown", "email", email)
		return nil, false, apperrors.ErrCodeCooldown
	}

	attempts, err := s.repo.IncrEmailAttempts(ctx, email)

	if err != nil {
		return nil, false, apperrors.ErrInternalError
	}

	if attempts > auth_repository.MaxAttempts {
		if attempts > auth_repository.MaxAttempts {
			slog.Warn("number of attempts exceeded", "email", email)
			return nil, false, apperrors.ErrTooManyAttempts
		}
	}

	savedCode, err := s.repo.GetCode(ctx, email)

	if err != nil {
		slog.Warn("error when getting the code from the repository", "email", email, "err", err)
		return nil, false, apperrors.ErrInvalidCode
	}

	if err := s.repo.SetCodeCooldown(ctx, email); err != nil {
		slog.Warn("error when setting code cooldown", "email", email)
		return nil, false, apperrors.ErrInternalError
	}

	if savedCode != code {
		slog.Warn("invalid code", "email", email)
		return nil, false, apperrors.ErrInvalidCode
	}

	s.repo.DeleteCode(ctx, email)
	s.repo.DeleteSession(ctx, sessionID)
	s.repo.DeletePending(ctx, sessionID)
	s.repo.ResetEmailAttempts(ctx, email)
	s.repo.ResetCodeAttempts(ctx, email)

	accessToken, err := auth_jwt.GenerateAccessToken(email, s.jwtSecret)
	if err != nil{
		slog.Warn("error when generate acces token", "email", email)
		return nil, false, apperrors.ErrInternalError
	}

	refreshToken, err := auth_jwt.GenerateRefreshToken(email, s.jwtSecret)
	if err != nil {
		slog.Warn("error when generate refresh token", "email", email)
		return nil, false, apperrors.ErrInternalError
	}

	if err := s.repo.SaveRefreshToken(ctx, refreshToken, email); err != nil {
		slog.Warn("error saving refresh token", "email", email)
		return nil, false, apperrors.ErrInternalError
	}

	exists, err := s.userRepo.UserExistsByEmail(ctx, email)
	if err != nil {
		slog.Warn("error checking the user existance's", "email", email, "err", err)
		return nil, false, apperrors.ErrInternalError
	}

	slog.Info("the user has been successfully logged in", "email", email)
	return &auth_models.TokenPair{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, !exists, nil
}

func (s *Service) GetMe(ctx context.Context, accessToken string) (*user_repository.User, error) {
	claims, err := auth_jwt.Parse(accessToken, s.jwtSecret)

	if err != nil {
		slog.Warn("invalid access token", "err", err)
		return nil, apperrors.ErrUnauthorized
	}

	if claims.TokenType != "access" {
		slog.Warn("wrong token type")
		return nil, apperrors.ErrUnauthorized
	}

	slog.Info("profile received", "email", claims.Email)

	return s.userRepo.GetUserByEmail(ctx, claims.Email)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*auth_models.TokenPair, error) {
	claims, err := auth_jwt.Parse(refreshToken, s.jwtSecret)
	if err != nil {
		slog.Warn("invalid refresh token", "err", err)
		return nil, apperrors.ErrUnauthorized
	}

	if claims.TokenType != "refresh" {
		slog.Warn("wrong token type", "token_type", claims.TokenType)
		return nil, apperrors.ErrUnauthorized
	}

	email, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		slog.Warn("refresh token not found in redis", "err", err)
		return nil, apperrors.ErrUnauthorized
	}

	if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		slog.Warn("error deleting refresh token", "err", err)
		return nil, apperrors.ErrInternalError
	}

	accessToken, err := auth_jwt.GenerateAccessToken(email, s.jwtSecret)
	if err != nil {
		slog.Warn("error when generate access roken", "err", err)
		return nil, apperrors.ErrInternalError
	}

	newRefreshToken, err := auth_jwt.GenerateRefreshToken(email, s.jwtSecret)
	if err != nil {
		slog.Warn("error when generate refresh token", "err", err)
		return nil, apperrors.ErrInternalError
	}

	if err := s.repo.SaveRefreshToken(ctx, newRefreshToken, email); err != nil {
		slog.Warn("error saving refresh token", "err", err)
		return nil, apperrors.ErrInternalError
	}

	slog.Info("tokon refreshed", "email", email)
	return &auth_models.TokenPair{
		RefreshToken: newRefreshToken,
		AccessToken: accessToken,
	}, nil
}

func (s *Service) Logout(ctx context.Context, token string) {
	s.repo.DeleteRefreshToken(ctx, token)
	slog.Info("the user goes out")
}

func (s *Service) CompleteRegistration(ctx context.Context, accessToken, nickname, name, status string) error {
	claims, err := auth_jwt.Parse(accessToken, s.jwtSecret)
	if err != nil {
		slog.Warn("invalid acces token")
		return apperrors.ErrUnauthorized
	}

	if claims.TokenType != "access" {
		slog.Warn("wrong token type", "token_type", claims.TokenType)
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

	if err = s.userRepo.SaveUserEmail(ctx, claims.Email, nickname, name, status, avatar_color); err != nil {
		slog.Error("error when saving the profile", "email", claims.Email, "err", err)
		return err
	}

	slog.Info("profile created", "email", claims.Email, "nickname", nickname)
	return nil
}
