package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/chat/chat_handler"
)

func SetupRoutes (engine *gin.Engine) {
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Test connection is running",
		})
	})
}

func DirectRouter (engine *gin.Engine, handler *chat_handler.DirectHandler) {
	chat := engine.Group("/chat/direct")
    chat.POST("", handler.CreateNewDirectHandler)
}