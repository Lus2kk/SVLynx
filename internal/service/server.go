package service

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth_handler"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/auth_service"
	"github.com/svlynx/messenger/internal/chat/chat_handler"
	"github.com/svlynx/messenger/internal/chat/chat_repository"
	"github.com/svlynx/messenger/internal/chat/chat_routes"
	"github.com/svlynx/messenger/internal/chat/chat_service"
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
	dsn := fmt.Sprintf(

		"postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Addr,
		cfg.Postgres.DB,
	)
	migration.RunMigrate(dsn)

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ReddisAddr,
	})

	emailSender := email.NewSender(cfg.SmtpHost, cfg.SmtpPort, cfg.SenderEmail, cfg.SenderPassword)

	


	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil{
		panic(err)
	}
	userRepo := user_repository.NewRepository(db)
	repo := auth_repository.NewRepository(redisClient)
	service := auth_service.NewService(repo, emailSender, userRepo, cfg.JWTSecret)
	handler := auth_handler.NewHandler(service, cfg.TelegramToken)

	postgresRepo := chat_repository.NewPostgresRepo(db)

	directService := chat_service.NewDirectService(postgresRepo)
	directHandler := chat_handler.NewDirectHandler(directService)

	messageService := chat_service.NewMessageService(postgresRepo)
	messageHandler := chat_handler.NewMessageHandler(messageService)

	Router := gin.Default()
	chat_routes.SetupRoutes(Router)
	chat_routes.DirectRouter(Router, directHandler)
	chat_routes.MessageRouter(Router, messageHandler)

	
	Router.Use(router.CorsMiddleware())
	router.RegisterRoutes(Router, handler)

	return &Server{
		cfg: cfg,
		router: Router,
		db: db,
	}
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.cfg.Port)
}

func (s *Server) Close() {
	s.db.Close()
}