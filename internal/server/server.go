package server

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	auth_handler "github.com/svlynx/messenger/internal/auth/handler"
	auth_repository "github.com/svlynx/messenger/internal/auth/repository"
	auth_routes "github.com/svlynx/messenger/internal/auth/routes"
	auth_service "github.com/svlynx/messenger/internal/auth/service"
	chat_handler "github.com/svlynx/messenger/internal/chat/handler"
	chat_repository "github.com/svlynx/messenger/internal/chat/repository"
	chat_routes "github.com/svlynx/messenger/internal/chat/routes"
	chat_service "github.com/svlynx/messenger/internal/chat/service"
	"github.com/svlynx/messenger/internal/chat/ws"
	"github.com/svlynx/messenger/internal/config"
	"github.com/svlynx/messenger/internal/middleware"
	"github.com/svlynx/messenger/internal/pkg/email"
	"github.com/svlynx/messenger/internal/push"
	user_repository "github.com/svlynx/messenger/internal/user/repository"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrate(PostgresAddr string) {
	
	m, err := migrate.New("file://migrations", PostgresAddr)
	if err != nil{ 
		panic(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
}

type Server struct {
	cfg        *config.Config
	router     *gin.Engine
	db         *pgxpool.Pool
	PushSender *push.Sender
}

func NewServer(cfg *config.Config) *Server {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Addr,
		cfg.Postgres.DB,
	)

	RunMigrate(dsn)

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ReddisAddr,
	})

	emailSender := email.NewSender(cfg)

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	userRepo := user_repository.NewRepository(db)
	repo := auth_repository.NewRepository(redisClient)
	service := auth_service.NewService(repo, emailSender, userRepo, cfg.JWTSecret)
	handler := auth_handler.NewHandler(service, cfg.TelegramToken)

	postgresRepo := chat_repository.NewPostgresRepo(db)
	directService := chat_service.NewDirectService(postgresRepo, userRepo)
	messageService := chat_service.NewMessageService(postgresRepo)

	pushRepo := push.NewRepository(db)
	pushHandler := push.NewHandler(pushRepo, cfg.JWTSecret)
	pushSender := push.NewSender(pushRepo, cfg.VAPIDPrivateKey, cfg.VAPIDPublicKey, cfg.VAPIDEmail)

	hub := ws.NewHub(messageService, directService)
	go hub.Run()

	messageHandler := chat_handler.NewMessageHandler(messageService, hub, pushSender)
	directHandler := chat_handler.NewDirectHandler(directService, hub)
	wsHandler := chat_handler.NewWsHandler(hub)

	Router := gin.Default()
	Router.Use(middleware.CorsMiddleware())

	push.RegisterRoutes(Router, pushHandler)

	chat_routes.SetupRoutes(Router)
	chat_routes.DirectRouter(Router, directHandler)
	chat_routes.MessageRouter(Router, messageHandler)
	chat_routes.WsRouter(Router, wsHandler)
	auth_routes.RegisterRoutes(Router, handler)

	return &Server{
		cfg:        cfg,
		router:     Router,
		db:         db,
		PushSender: pushSender,
	}
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.cfg.Port)
}

func (s *Server) Close() {
	s.db.Close()
}
