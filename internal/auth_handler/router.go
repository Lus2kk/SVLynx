package auth_handler

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler){
	authTg := r.Group("/auth/telegram")
	authTg.POST("/init", h.InitTelegramAuth())

	emailAuth := r.Group("/auth/email")
    emailAuth.POST("/init", h.InitEmailAuth())  
    emailAuth.POST("/send-code", h.SendEmailCode())  
    emailAuth.POST("/verify-code", h.VerifyEmailCode()) 
	emailAuth.POST("/complete", h.CompleteRegistration())
	emailAuth.POST("/logout", h.Logout())
	emailAuth.GET("/me", h.GetMe())
}