package auth_handler

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler){
	auth := r.Group("/auth/telegram")
	auth.POST("/init", h.InitTelegramAuth())
	auth.POST("/callback", h.TelegramCallback())
}


