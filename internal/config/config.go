package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
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
