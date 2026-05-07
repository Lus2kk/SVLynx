package push

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/svlynx/messenger/internal/auth_jwt"
)

type Handler struct {
	repo *Repository
	jwtSecret string
}

func NewHandler(r *Repository, jwtSecret string) *Handler {
	return &Handler{
		repo: r,
		jwtSecret: jwtSecret,
	}
}
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}
		claims, err := auth_jwt.Parse(parts[1], h.jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userID", claims.Subject)
		c.Next()
	}
}
type subscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
	P256dh   string `json:"p256dh" binding:"required"`
	Auth     string `json:"auth" binding:"required"`
}

func (h *Handler) Subscribe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sub := PushSubscription{
		UserID:    userID.(string),
		Endpoint:  req.Endpoint,
		P256dh:    req.P256dh,
		Auth:      req.Auth,
		UserAgent: c.GetHeader("User-Agent"),
	}
	if err := h.repo.Save(c.Request.Context(), sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to save subscription"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type unsubscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
}

func (h *Handler) Unsubscribe(c *gin.Context) {
	var req unsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), req.Endpoint); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
