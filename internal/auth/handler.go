package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Hanlder struct {
	service *Service
}

func NewHandler(service *Service) *Hanlder {
	return &Hanlder{service: service}
}

func (h *Hanlder) InitTelegramAuth(c *gin.Context) {
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
