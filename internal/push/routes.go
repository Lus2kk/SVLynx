package push

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler) {
	group := r.Group("/push", h.AuthMiddleware())
	{
		group.POST("/subscribe", h.Subscribe)
		group.POST("/unsubscribe", h.Unsubscribe)
	}
}