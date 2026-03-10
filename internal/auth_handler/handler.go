package auth_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/auth_service"
)

type Handler struct {
	service *auth_service.Service
}

func NewHandler(service *auth_service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitTelegramAuth() gin.HandlerFunc {
	return func(c *gin.Context){
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
