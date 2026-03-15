package main

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth_handler"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/auth_service"
	"github.com/svlynx/messenger/internal/config"
	"github.com/gin-contrib/cors"
)

func main(){
	cfg := config.MustLoad()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ReddisAddr,
	})

	repo := auth_repository.NewRepository(redisClient)
	service := auth_service.NewService(repo,cfg.TelegramToken)
	handler := auth_handler.NewHandler(service)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"POST", "GET", "OPTIONS"},
    AllowHeaders: []string{"Content-Type"},
}))

	auth_handler.RegisterRoutes(r, handler)
	r.Run(":8080")
}
