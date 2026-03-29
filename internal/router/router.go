	package router

	import (
		"net/http"

		"github.com/gin-gonic/gin"
		"github.com/svlynx/messenger/internal/auth_handler"
		"github.com/svlynx/messenger/internal/chat/chat_handler"
		"github.com/svlynx/messenger/internal/chat/ws"
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
	}

	func MessageRouter(engine *gin.Engine, handler *chat_handler.MessageHandler) {
		message := engine.Group("/chat/messages")
		message.POST("", handler.SendMessageHandler)
		message.GET("", handler.GetMessagesByChatIdHandler)
		message.GET("/search", handler.SearchMessageHandler)
		message.PATCH("/:id/status", handler.UpdateMessageStatusHandler)
		message.DELETE("/:id", handler.DeleteMessageHandler)
	}

	func WebsocketChatRoutes (engine *gin.Engine, handler *ws.WsHandler) {
		engine.GET("/ws", handler.ConnectionHandler)
	}