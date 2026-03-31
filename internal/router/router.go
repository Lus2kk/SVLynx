package router

import (
	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/auth_handler"
)

func RegisterRoutes(r *gin.Engine, h *auth_handler.Handler){
	auth := r.Group("/auth")
	auth.POST("/refresh", h.Refresh())
	auth.POST("/logout", h.Logout())
	auth.POST("/me", h.GetMe())
	
	telegram := auth.Group("/telegram")
	telegram.POST("/init", h.InitEmailAuth())
	telegram.POST("/callback", h.TelegramCallback())

	email:= auth.Group("/email")
    email.POST("/init", h.InitEmailAuth())  
    email.POST("/send-code", h.SendEmailCode())  
    email.POST("/verify-code", h.VerifyEmailCode())
	email.POST("/complete", h.CompleteRegistration())
}