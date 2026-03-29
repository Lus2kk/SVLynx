package router

import (
	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/auth_handler"
)

func RegisterRoutes(r *gin.Engine, h *auth_handler.Handler){
	authTg := r.Group("/auth/telegram")
	authTg.POST("/init", h.InitTelegramAuth())
	authTg.POST("/callback", h.TelegramCallback())

	emailAuth := r.Group("/auth/email")
    emailAuth.POST("/init", h.InitEmailAuth())  
    emailAuth.POST("/send-code", h.SendEmailCode())  
    emailAuth.POST("/verify-code", h.VerifyEmailCode())
	emailAuth.POST("/refresh", h.Refresh()) 
	emailAuth.POST("/complete", h.CompleteRegistration())
	emailAuth.POST("/logout", h.Logout())
	emailAuth.GET("/me", h.GetMe())
}