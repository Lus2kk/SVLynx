package main

import (
	"fmt"
	"log"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth_handler"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/auth_service"
	"github.com/svlynx/messenger/internal/chat/chat_handler"
	"github.com/svlynx/messenger/internal/chat/chat_repository"
	"github.com/svlynx/messenger/internal/chat/chat_routes"
	"github.com/svlynx/messenger/internal/chat/chat_service"
	"github.com/svlynx/messenger/internal/config"
	"github.com/svlynx/messenger/internal/repository"
)

func main() {
	cfg := config.MustLoad()
	//redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ReddisAddr,
	})
	//auth repo
	repo := auth_repository.NewRepository(redisClient)
	service := auth_service.NewService(repo, cfg.TelegramToken)
	handler := auth_handler.NewHandler(service)

	///chat initialization
	dsn := fmt.Sprintf(

		"postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Addr,
		cfg.Postgres.DB,
	)
	db, err := repository.NewDB(dsn)
	if err != nil {
		log.Fatal("have no connection to database! ", err.Error())
	}
	postgresRepo := chat_repository.NewPostgresRepo(db)

	directService := chat_service.NewDirectService(postgresRepo)
	directHandler := chat_handler.NewDirectHandler(directService)

	messageService := chat_service.NewMessageService(postgresRepo)
	messageHandler := chat_handler.NewMessageHandler(messageService)

	///router
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "GET", "OPTIONS"},
		AllowHeaders: []string{"Content-Type"},
	}))

	///routes
	auth_handler.RegisterRoutes(router, handler)
	chat_routes.SetupRoutes(router)
	chat_routes.DirectRouter(router, directHandler)
	chat_routes.MessageRouter(router, messageHandler)

	router.Run(":8080")
}
