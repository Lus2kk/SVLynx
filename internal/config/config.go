package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	PostgresAddr   string `env:"POSTGRESQL_ADDR" env-required:"true"` 
	SmtpHost       string `env:"SMTP_HOST"     env-default:"smtp.gmail.com"`
    SmtpPort       int    `env:"SMTP_PORT"     env-default:"587"`
	SenderEmail	   string `env:"SENDER_EMAIL" env-required:"true"`
	SenderPassword string `env:"SENDER_PASSWORD" env-required:"true"`
	JWTSecret	   string `env:"JWT_SECRET" env-reqiured:"true"`

	TelegramToken string `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
	ReddisAddr    string `env:"REDIS_ADDR" env-required:"true"`
	Port          string `env:"PORT" env-required:"true"`
}

func MustLoad() *Config {
	cfg := &Config{}
	if err := cleanenv.ReadConfig(".env", cfg); err != nil {
		panic(err)
	}
	return cfg
}
