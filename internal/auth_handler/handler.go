package auth_handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/apperrors"
	"github.com/svlynx/messenger/internal/auth_models"
	"github.com/svlynx/messenger/internal/auth_service"
)

type Handler struct {
	service       *auth_service.Service
	telegramToken string
}

type SendCodeDTO struct {
	SessionID string `json:"session_id" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

type VerifyCodeDTO struct {
	SessionID string `json:"session_id" binding:"required"`
	Code      string `json:"code" binding:"required,len=6"`
}

type CompleteRegistrationDTO struct {
	Nickname string `json:"nickname" binding:"required,min=3,max=25"`
	Name     string `json:"name"`
	Status   string `json:"status"`
}

func HandlerError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrSessionNotFound):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrCodeExpired):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrInvalidCode):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrEmailCooldown):
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrCodeCooldown):
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrTooManyAttempts):
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrNicknameExists):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrSessionCreate):
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrSaveUserFailed):
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrEmailSendFailed):
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrInvalidHash):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error()},
		)
	case errors.Is(err, apperrors.ErrAuthExpired):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
	case errors.Is(err, apperrors.ErrInvalidRequest):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": apperrors.ErrInternal.Error(),
		})
	}
}

func NewHandler(service *auth_service.Service, telegramToken string) *Handler {
	return &Handler{
		service: service,
	telegramToken: telegramToken,
}}


func (h *Handler) InitTelegramAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session, err := h.service.InitSession(ctx)
		if err != nil {
			HandlerError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"session_id": session.SessionID,
			"expires_at": session.ExpiresAt,
		})
	}
}

func (h *Handler) getBearer(c *gin.Context) string {
	header := c.GetHeader("Authorization")

	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		 slog.Warn("invalid auth header format", "value", header)
		return ""
	}
	
	return parts[1]
}
 
func (h *Handler) InitEmailAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		session, err := h.service.InitSession(ctx)
		if err != nil {
			HandlerError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"session_id": session.SessionID,
			"expires_at": session.ExpiresAt,
		})

	}
}

func (h *Handler) SendEmailCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var req SendCodeDTO

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "the email address was entered incorrectly",
			})
			return
		}

		if err := h.service.SendConfirmationCode(ctx, req.SessionID, req.Email); err != nil {
			HandlerError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "The code has been sent to " + req.Email,
		})
	}
}

func (h *Handler) VerifyEmailCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var req VerifyCodeDTO

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "enter the 6-digit code"})
			return
		}

		tokens, isNew, err := h.service.VerifyCode(ctx, req.SessionID, req.Code)

		if err != nil {
			HandlerError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token": tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
			"is_new": isNew, // для complete reg
		})
	}
}

func (h *Handler) Refresh() gin.HandlerFunc {
	return func (c *gin.Context) {
		ctx := c.Request.Context()
		
		refreshToken := c.GetHeader("X-Refresh-Token")
		if refreshToken == "" {
			HandlerError(c, apperrors.ErrUnauthorized)
			return 
		}

		tokens, err := h.service.Refresh(ctx, refreshToken)
		if err != nil {
			HandlerError(c, err)
			return 
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token": tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		})
	}
}

func (h *Handler) GetMe() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		accessToken := h.getBearer(c)
		if accessToken == "" {
			HandlerError(c, apperrors.ErrUnauthorized)
			return 
		}

		user, err := h.service.GetMe(ctx, accessToken)
		if err != nil {
			HandlerError(c, err)
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func (h *Handler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		refreshToken := c.GetHeader("X-Refresh-Token")
		if refreshToken != "" {
			h.service.Logout(ctx, refreshToken)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "logged out",
		})
	}
}

func (h *Handler) CompleteRegistration() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var req CompleteRegistrationDTO

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		accessToken := h.getBearer(c)

		if accessToken == "" {
			HandlerError(c, apperrors.ErrUnauthorized)
			return
		}

		if err := h.service.CompleteRegistration(ctx, accessToken, req.Nickname, req.Name, req.Status); err != nil {
			HandlerError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "profile created",
		})
	}
}

func (h *Handler) TelegramCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var req auth_models.TelegramCallbackRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			HandlerError(c, apperrors.ErrInvalidRequest)
			return
		}

		if err := validateTelegramCallback(&req); err != nil {
			HandlerError(c, err)
			return 
		}
		tokens, err := h.service.TelegramCallback(ctx, h.telegramToken, &req)
		if err != nil {
			HandlerError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"access_token": tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		})
	}
}
