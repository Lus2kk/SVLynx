package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
    ResendAPIKey   string `env:"RESEND_API_KEY"  env-required:"true"`
    ResendFrom     string `env:"RESEND_FROM"     env-default:"onboarding@resend.dev"`
    JWTSecret      string `env:"JWT_SECRET"      env-required:"true"`
    Port           string `env:"PORT"            env-required:"true"`
    TelegramToken  string `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
    ReddisAddr     string `env:"REDIS_ADDR"      env-required:"true"`
    Postgres       PostgresConfig
}

type PostgresConfig struct {
    User     string `env:"POSTGRES_USER"     env-required:"true"`
    Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
    Addr     string `env:"POSTGRES_ADDR"     env-required:"true"`
    DB       string `env:"POSTGRES_DB"       env-required:"true"`
}

func MustLoad() *Config {
    cfg := &Config{}
    if err := cleanenv.ReadConfig(".env", cfg); err != nil {
        panic(err)
    }
    return cfg
}