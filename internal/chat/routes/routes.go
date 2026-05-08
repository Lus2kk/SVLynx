package chat_routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	chat_handler "github.com/svlynx/messenger/internal/chat/handler"
)

func SetupRoutes(engine *gin.Engine) {
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Test connection is running",
		})
	})
}

func DirectRouter(engine *gin.Engine, handler *chat_handler.DirectHandler) {
	chat := engine.Group("/chat/direct")
	chat.POST("", handler.CreateNewDirectHandler)
	chat.GET("", handler.GetDirectByIdHandler)
	chat.GET("/list", handler.GetListOfDirectsByIDHandler)
	chat.DELETE("/:id", handler.DeleteDirectHandler)

	users := engine.Group("/users")
    users.GET("/search", handler.SearchUsersHandler)
	users.GET("/:id/status", handler.GetUserStatusHandler)
}

func MessageRouter(engine *gin.Engine, handler *chat_handler.MessageHandler) {
	message := engine.Group("/chat/messages")
	message.POST("", handler.SendMessageHandler)
	message.POST("/voice", handler.SendVoiceMessageHandler)
	message.GET("", handler.GetMessagesByChatIdHandler)
	message.GET("/search", handler.SearchMessageHandler)
	message.PATCH("/:id/status", handler.UpdateMessageStatusHandler)
	message.PATCH("/read", handler.MarkChatMessagesAsReadHandler)
	message.DELETE("/:id", handler.DeleteMessageHandler)
}

func WsRouter(engine *gin.Engine, handler *chat_handler.WsHandler) {
	engine.GET("/ws", handler.ServeWs)
}
