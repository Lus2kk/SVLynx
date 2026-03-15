package auth_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/auth_models"
	"github.com/svlynx/messenger/internal/auth_service"
)

type Handler struct {
	service *auth_service.Service
}

func NewHandler(service *auth_service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitTelegramAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := h.service.InitSession()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Session creation error",
			})
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

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}
		session, err := h.service.TelegramCallback(&req)
		if err != nil {
			switch err.Error() {
			case "invalid hash":
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid hash"})
			case "auth data expired":
				c.JSON(http.StatusUnauthorized, gin.H{"error": "auth data expired"})
			case "session not found":
				c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
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
