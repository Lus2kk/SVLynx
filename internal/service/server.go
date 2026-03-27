package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth_handler"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/auth_service"
	"github.com/svlynx/messenger/internal/config"
	"github.com/svlynx/messenger/internal/email"
	"github.com/svlynx/messenger/internal/migration"
	"github.com/svlynx/messenger/internal/router"
	"github.com/svlynx/messenger/internal/user_repository"
)

type Server struct {
	cfg    *config.Config
	router *gin.Engine
	db 	   *pgxpool.Pool
}

func NewServer(cfg *config.Config) *Server {
	migration.RunMigrate(cfg.PostgresAddr)

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ReddisAddr,
	})

	emailSender := email.NewSender(cfg.SmtpHost, cfg.SmtpPort, cfg.SenderEmail, cfg.SenderPassword)

	db, err := pgxpool.New(context.Background(), cfg.PostgresAddr)
	if err != nil{
		panic(err)
	}

	userRepo := user_repository.NewRepository(db)

	repo := auth_repository.NewRepository(redisClient)

	service := auth_service.NewService(repo, emailSender, userRepo, cfg.JWTSecret)

	handler := auth_handler.NewHandler(service, cfg.TelegramToken)

	r := gin.Default()
	r.Use(router.CorsMiddleware())
	router.RegisterRoutes(r, handler)

	return &Server{
		cfg: cfg,
		router: r,
		db: db,
	}
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.cfg.Port)
}

func (s *Server) Close() {
	s.db.Close()
}