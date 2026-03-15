package auth_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/auth_service"
)

type Handler struct {
	service *auth_service.Service
}

type SendCodeRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

type VerifyCodeRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Code      string `json:"code" binding:"required,len=6"`
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

func (h *Handler) InitEmailAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
	session, err := h.service.InitSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ошибка при создании сессии",
		})
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
		var req SendCodeRequest

		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Почта введена неверно",
			})
			return
		}

		if err := h.service.SendConfirmationCode(req.SessionID, req.Email); err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Ошибка при отправке кода на почту",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Код отправлен на " + req.Email,
		})
	}
}

func (h *Handler) VerifyEmailCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyCodeRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := h.service.VerifyCode(req.SessionID, req.Code); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "email подтвержден",
		})
	}
}
