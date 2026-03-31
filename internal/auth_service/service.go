package auth_service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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
		return nil, err
	}
	return &auth_models.Session{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(auth_repository.SessionTTL),
		Status:    auth_models.StatusPending,
	}, nil
}

func (s *Service) SendConfirmationCode(ctx context.Context, sessionID, email string) error {
	ok, err := s.repo.SessionExists(ctx, sessionID)
    if err != nil {
       slog.Warn("error when check session existence")
	   return apperrors.ErrSessionCreate
    }

	if !ok {
		slog.Warn("session not exists")
		return apperrors.ErrSessionNotFound
	}

	onCooldown, err := s.repo.EmailCooldownExists(ctx, email)

	if err != nil {
		return apperrors.ErrInternal
	}

	if onCooldown {
		slog.Warn("blocked by email cooldown", "email", email)
		return apperrors.ErrEmailCooldown
	}

	attempts, err := s.repo.IncrEmailAttempts(ctx, email)

	if err != nil {
		return apperrors.ErrInternal
	}

	if attempts > auth_repository.MaxAttempts {
		slog.Warn("number of attempts exceeded", "email", email)
		return apperrors.ErrTooManyAttempts
	}

	code, err := auth_code.GenerateSixDigitCode()

	if err != nil {
		return apperrors.ErrInternal
	}

	err = s.repo.SavePending(ctx, sessionID, email)

	if err != nil {
		return apperrors.ErrInternal
	}

	err = s.repo.SaveCode(ctx, email, code)

	if err != nil {
		return apperrors.ErrInternal
	}

	if err := s.repo.SetEmailCooldown(ctx, email); err != nil {
		slog.Warn("error when setting email cooldown", "email", email)
		return apperrors.ErrInternal
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
	if err != nil {
    return nil, false, apperrors.ErrInternal
	}

	if onCooldown {
		slog.Warn("blocked by code cooldown", "email", email)
		return nil, false, apperrors.ErrCodeCooldown
	}

	attempts, err := s.repo.IncrCodeAttempts(ctx, email)

	if err != nil {
		return nil, false, apperrors.ErrInternal
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
		return nil, false, apperrors.ErrInternal
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

	exists, err := s.userRepo.UserExistsByEmail(ctx, email)
	if err != nil {
		slog.Warn("error checking the user existance", "email", email, "err", err)
		return nil, false, apperrors.ErrInternal
	}

	if !exists {
		if err := s.userRepo.SaveUserEmail(ctx, email, "", "", "", ""); err != nil {
			slog.Warn("error when save the user profile", "email", email, "err", err)
			return nil, false, apperrors.ErrInternal
		}
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)

	accessToken, err := auth_jwt.GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil{
		slog.Warn("error when generate acces token", "email", email)
		return nil, false, apperrors.ErrInternal
	}

	refreshToken, err := auth_jwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		slog.Warn("error when generate refresh token", "email", email)
		return nil, false, apperrors.ErrInternal
	}

	if err := s.repo.SaveRefreshToken(ctx, refreshToken, user.ID); err != nil {
		slog.Warn("error saving refresh token", "user_id", user.ID)
		return nil, false, apperrors.ErrInternal
	}

	slog.Info("the user has been successfully logged in", "user_id", user.ID)

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

	userID := claims.Subject

	if userID == "" {
		slog.Warn("empty subject in access token")
        return nil, apperrors.ErrUnauthorized
	}

	slog.Info("profile received", "user_id", userID)

	return s.userRepo.GetUserByUserID(ctx, claims.Subject)
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

	userID, err := s.repo.GetUserIDByRefreshToken(ctx, refreshToken)

	if err != nil {
		slog.Warn("refresh token not found in redis", "err", err)
		return nil, apperrors.ErrUnauthorized
	}

	if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		slog.Warn("error deleting refresh token", "err", err)
		return nil, apperrors.ErrInternal
	}

	accessToken, err := auth_jwt.GenerateAccessToken(userID, s.jwtSecret)
	if err != nil {
		slog.Warn("error when generate access roken", "err", err)
		return nil, apperrors.ErrInternal
	}

	newRefreshToken, err := auth_jwt.GenerateRefreshToken(userID, s.jwtSecret)
	if err != nil {
		slog.Warn("error when generate refresh token", "err", err)
		return nil, apperrors.ErrInternal
	}

	if err := s.repo.SaveRefreshToken(ctx, newRefreshToken, userID); err != nil {
		slog.Warn("error saving refresh token", "err", err)
		return nil, apperrors.ErrInternal
	}

	slog.Info("token refreshed", "user_id", userID)
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

	userID := claims.Subject
    if userID == "" {
        slog.Warn("empty subject in access token")
        return apperrors.ErrUnauthorized
    }
	
	exists, err := s.userRepo.NicknameExists(ctx, nickname)
	if err != nil {
		slog.Warn("error checking the existence of nickname", "nickname", nickname, "err", err)
		return apperrors.ErrInternal
	}

	if exists {
		slog.Warn("username already exists", "nickname", nickname)
		return apperrors.ErrNicknameExists
	}

	colors := []string{"#2a2379", "#1D9E75", "#D85A30", "#378ADD", "#b51ed7"}
	avatar_color := colors[rand.Intn(len(colors))]

	if status == "" {
		status = "Привет!"
	}

	if err = s.userRepo.UpdateUserProfile(ctx, userID, nickname, name, status, avatar_color); err != nil {
		slog.Error("error when saving the profile", "user_id", userID, "err", err)
		return err
	}

	slog.Info("profile created", "user_id", userID, "nickname", nickname)
	return nil
}

func (s *Service) TelegramCallback(ctx context.Context, telegramToken string, req *auth_models.TelegramCallbackRequest) (*auth_models.TokenPair, error) {
	if !s.verifyHash(telegramToken, req) {
		slog.Warn("error invalid hash", "telegram_id", req.ID)
		return nil, apperrors.ErrInvalidHash
	}

	if time.Now().Unix()-req.AuthDate > 86400 {
		slog.Warn("auth data expired", "telegram_id", req.ID, "auth_date", req.AuthDate)
		return nil, apperrors.ErrAuthExpired
	}

	exists, err := s.userRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		slog.Warn("error when check username existance", "telegram_id", req.ID, "err", err)
		return nil, apperrors.ErrInternal
	}

	if !exists {
		if err := s.userRepo.SaveUserTelegram(
			ctx,
			req.ID,
			req.Username, 
			req.FirstName,
			req.LastName,
			req.PhotoURL); err != nil {
				slog.Warn("error when save telegram user", "telegram_id", req.ID, "err", err)
				return nil, apperrors.ErrInternal
			}
	}

	user, err := s.userRepo.GetUserByTgID(ctx, req.ID)
	if err != nil {
		slog.Warn("error when get telegram user", "telegram_id", req.ID, "err", err)
		return nil, apperrors.ErrInternal
	}

	accessToken, err := auth_jwt.GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		slog.Warn("error when generate access token", "user_id", user.ID ,"err", err)
		return nil, apperrors.ErrInternal
	}

	refreshToken, err := auth_jwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		slog.Warn("error when generate refresh token", "user_id", user.ID ,"err", err)
		return nil, apperrors.ErrInternal
	}
	
	if err := s.repo.SaveRefreshToken(ctx, refreshToken, user.ID); err != nil {
		slog.Warn("error when save refresh token","user_id", user.ID, "err", err)
		return nil, apperrors.ErrInternal
	}

	return &auth_models.TokenPair{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

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
	if req.LastName != "" {
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

