package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	ReddisAddr    string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("")
	}
	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ReddisAddr:    os.Getenv("REDIS_ADDR"),
	}
}
