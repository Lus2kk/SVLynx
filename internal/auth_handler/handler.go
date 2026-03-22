package auth_handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/apperrors"
	"github.com/svlynx/messenger/internal/auth_service"
)

type Handler struct {
	service *auth_service.Service
}

type SendCodeDTO struct {
	SessionID string `json:"session_id" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

type VerifyCodeDTO struct {
	SessionID string `json:"session_id" binding:"required"`
	Code      string `json:"code" binding:"required,len=6"`
}

type CompleteRegistrationDTO struct{ 
	Nickname string `json:"nickname" binding:"required,min=3,max=25"`
	Name 	 string `json:"name"`
	Status 	 string `json:"status"`
}

func HandlerError(c *gin.Context, err error){
	switch{
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
	case errors.Is(err, apperrors.ErrCooldown):
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
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": apperrors.ErrInternalError.Error(),
		})
	}
}

func NewHandler(service *auth_service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitTelegramAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		session, err := h.service.InitSession(ctx)
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
	ctx := context.Background()
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
		ctx := context.Background()
		var req SendCodeDTO

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "почта введена неверно"})
			return
		}

		if err := h.service.SendConfirmationCode(ctx, req.SessionID, req.Email); err != nil {
			HandlerError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Код отправлен на " + req.Email,
		})
	}
}

func (h *Handler) VerifyEmailCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		var req VerifyCodeDTO

		if err := c.ShouldBindJSON(&req); err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "введите 6-значный код"})
			return
		}

		token, isNew, err := h.service.VerifyCode(ctx, req.SessionID, req.Code)
		
		if err != nil {
			HandlerError(c, err)
			return
		}

		c.SetCookie("auth_token", token, 30*24*3600, "/", "", false, true)

		c.JSON(http.StatusOK, gin.H{
			"is_new": isNew, // для регистрации
		})
	}
}

func (h *Handler) GetMe() gin.HandlerFunc {
	return func (c *gin.Context) {
		ctx := context.Background()
		token, err := c.Cookie("auth_token")
		if err != nil{
			HandlerError(c, apperrors.ErrUnauthorized)
			return 
		}

		user, err := h.service.GetMe(ctx, token)
		if err != nil{
			HandlerError(c, err)
			return 
		}

		c.JSON(http.StatusOK, user)
	}
}

func (h *Handler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		token, err := c.Cookie("auth_token")
		if err == nil {
			h.service.Logout(ctx, token)
		}

		c.SetCookie("auth_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{
			"message": "вышел",
		})
	}
}

func (h *Handler) CompleteRegistration() gin.HandlerFunc{
	return func (c *gin.Context) {
		ctx := context.Background()
		var req CompleteRegistrationDTO

		if err := c.ShouldBindJSON(&req); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return 
		}

		token, err := c.Cookie("auth_token")
		if err != nil{
			HandlerError(c, apperrors.ErrUnauthorized)
			return 
		}

		if err := h.service.CompleteRegistration(ctx, token, req.Nickname, req.Name, req.Status); err != nil{
			HandlerError(c, err)
			return 
		}

		c.JSON(http.StatusOK, gin.H{"message": "профиль создан"})
	}
}