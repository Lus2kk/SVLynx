package main

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth"
	"github.com/svlynx/messenger/internal/config"
)

func main(){
	cfg := config.Load()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ReddisAddr,
	})

	repo := auth.NewRepository(redisClient)
	service := auth.NewService(repo)
	handler := auth.NewHandler(service)

	r := gin.Default()

	r.POST("auth/telegram/init", handler.InitTelegramAuth)
	r.Run(":8080")
}
