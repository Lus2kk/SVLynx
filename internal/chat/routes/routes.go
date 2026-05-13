package chat_routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	channel_handler "github.com/svlynx/messenger/internal/chat/channel/handler"
	chat_handler "github.com/svlynx/messenger/internal/chat/direct/handler"
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
	message.POST("/media", handler.SendMediaMessageHandler)
	message.GET("", handler.GetMessagesByChatIdHandler)
	message.GET("/search", handler.SearchMessageHandler)
	message.PATCH("/:id/status", handler.UpdateMessageStatusHandler)
	message.PATCH("/read", handler.MarkChatMessagesAsReadHandler)
	message.DELETE("/:id", handler.DeleteMessageHandler)
}

func WsRouter(engine *gin.Engine, handler *chat_handler.WsHandler) {
	engine.GET("/ws", handler.ServeWs)
}


func RegisterChannelRoutes(engine *gin.Engine, h *channel_handler.ChannelHandler) {
	

	channels := engine.Group("/channels")
	{
		channels.POST("", h.CreateChannelHandler)           
		channels.GET("", h.GetUserChannelsHandler)          
		channels.GET("/search", h.SearchChannelsHandler)    
		channels.GET("/handle/:handle", h.GetChannelByHandleHandler)

		channels.GET("/:id", h.GetChannelByIDHandler)       
		channels.PATCH("/:id", h.UpdateChannelHandler)      
		channels.DELETE("/:id", h.DeleteChannelHandler)     

	
		channels.POST("/:id/join", h.JoinChannelHandler)   
		channels.POST("/:id/leave", h.LeaveChannelHandler)  
		channels.GET("/:id/members", h.GetMembersHandler)   

		channels.DELETE("/:id/members/:user_id", h.KickMemberHandler)              
		channels.PATCH("/:id/members/:user_id/role", h.UpdateMemberRoleHandler)    
		channels.POST("/:id/transfer", h.TransferOwnershipHandler)                

		
		channels.POST("/:id/posts", h.CreatePostHandler)               
		channels.GET("/:id/posts", h.GetPostsHandler)                 
		channels.GET("/:id/posts/pinned", h.GetPinnedPostsHandler)      
		channels.GET("/:id/posts/search", h.SearchPostsHandler)         

		channels.PATCH("/:id/posts/:post_id", h.UpdatePostHandler)      
		channels.DELETE("/:id/posts/:post_id", h.DeletePostHandler)     
		channels.PATCH("/:id/posts/:post_id/pin", h.PinPostHandler)    
		channels.POST("/:id/posts/:post_id/view", h.ViewPostHandler)   

		channels.POST("/:id/invites", h.CreateInviteLinkHandler)       
		channels.GET("/:id/invites", h.GetInviteLinksHandler)          
		channels.DELETE("/:id/invites/:link_id", h.DeactivateInviteLinkHandler) 
	}

	
	engine.POST("/invites/:token/join", h.JoinByInviteHandler)
}