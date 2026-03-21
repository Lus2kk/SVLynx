package auth_handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/auth_service"
	"github.com/svlynx/messenger/internal/config"
)

type Server struct {
	cfg    *config.Config
	router *gin.Engine
}

func NewServer(cfg *config.Config) *Server {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ReddisAddr,
	})

	repo := auth_repository.NewRepository(redisClient)
	service := auth_service.NewService(repo)
	handler := NewHandler(service, cfg.TelegramToken)

	r := gin.Default()
	r.Use(CorsMiddleware())
	RegisterRoutes(r, handler)

	return &Server{cfg: cfg, router: r}
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.cfg.Port)
}