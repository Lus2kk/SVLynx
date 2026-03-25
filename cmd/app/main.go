package main

import (
	"context"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/svlynx/messenger/internal/auth_handler"
	"github.com/svlynx/messenger/internal/auth_repository"
	"github.com/svlynx/messenger/internal/auth_service"
	"github.com/svlynx/messenger/internal/config"
	"github.com/svlynx/messenger/internal/email"
	"github.com/svlynx/messenger/internal/migration"
	"github.com/svlynx/messenger/internal/user_repository"
)



func main(){
	cfg := config.MustLoad()

	migration.RunMigrate(cfg.PostgresAddr)
	
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.ReddisAddr,
	})

	db, err := pgxpool.New(context.Background(), cfg.PostgresAddr)
	if err != nil{
		panic(err)
	}
	defer db.Close()

	repo := auth_repository.NewRepository(redisClient)
	userRepo := user_repository.NewRepository(db)

	emailSender := email.NewSender(cfg.SmtpHost, cfg.SmtpPort, cfg.SenderEmail, cfg.SenderPassword)

	service := auth_service.NewService(repo, emailSender, userRepo, cfg.JWTSecret)
	
	handler := auth_handler.NewHandler(service)

	r := gin.Default()

	auth_handler.RegisterRoutes(r, handler)
	r.Run(":8080")
}
