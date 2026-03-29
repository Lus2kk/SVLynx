package auth_handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/apperrors"
	"github.com/svlynx/messenger/internal/auth_models"
	"github.com/svlynx/messenger/internal/auth_service"
)

type Handler struct {
	service       *auth_service.Service
	telegramToken string
}

func NewHandler(service *auth_service.Service, telegramToken string) *Handler {
	return &Handler{
		service:       service,
		telegramToken: telegramToken}
}

func handlerError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidHash):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrAuthExpired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrInvalidRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrInternal})
	}
}

func (h *Handler) InitTelegramAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		session, err := h.service.InitSession(ctx)
		if err != nil {
			handlerError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"session_id": session.SessionID,
			"expires_at": session.ExpiresAt,
		})
	}
}

func (h *Handler) TelegramCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth_models.TelegramCallbackRequest

		ctx := c.Request.Context()

		if err := c.ShouldBindJSON(&req); err != nil {
			handlerError(c, apperrors.ErrInvalidRequest)
			return
		}
		
		if err := validateTelegramCallback(&req); err != nil {
			handlerError(c, err)
		}
		session, err := h.service.TelegramCallback(ctx, h.telegramToken, &req)
		if err != nil {
			handlerError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":      session.Status,
			"session_id":  session.SessionID,
			"first_name":  session.FirstName,
			"username":    session.Username,
			"telegram_id": session.TelegramID,
			"photo_url":   session.PhotoURL,
		})
	}
}
