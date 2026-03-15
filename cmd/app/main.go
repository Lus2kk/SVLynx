package main

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth_handler"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/auth_service"
	"github.com/svlynx/messenger/internal/config"
	"github.com/svlynx/messenger/internal/email"
)

func main(){
	cfg := config.MustLoad()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ReddisAddr,
	})

	repo := auth_repository.NewRepository(redisClient)

	emailSender := email.NewSender(cfg.SmtpHost, cfg.SmtpPort, cfg.SenderEmail, cfg.SenderPassword)

	service := auth_service.NewService(repo, emailSender)
	
	handler := auth_handler.NewHandler(service)

	r := gin.Default()

	auth_handler.RegisterRoutes(r, handler)
	r.Run(":8080")
}
