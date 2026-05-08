package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	AppEnv         string `env:"APP_ENV"          env-default:"development"`
	SmtpHost       string `env:"SMTP_HOST"        env-default:"smtp.gmail.com"`
	SmtpPort       int    `env:"SMTP_PORT"        env-default:"587"`
	SmtpEmail      string `env:"SMTP_EMAIL"`
	SmtpPassword   string `env:"SMTP_PASSWORD"`
	ResendApiKey   string `env:"RESEND_API_KEY"`
	SenderEmail    string `env:"SENDER_EMAIL"`
	JWTSecret      string `env:"JWT_SECRET" env-required:"true"`
	Port           string `env:"PORT" env-required:"true"`
	TelegramToken  string `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
	ReddisAddr     string `env:"REDIS_ADDR" env-required:"true"`
	Postgres       PostgresConfig
	VAPIDPublicKey  string `env:"VAPID_PUBLIC_KEY"  env-required:"true"`
    VAPIDPrivateKey string `env:"VAPID_PRIVATE_KEY" env-required:"true"`
    VAPIDEmail      string `env:"VAPID_EMAIL"       env-required:"true"`
}
type PostgresConfig struct {
	User     string `env:"POSTGRES_USER"        env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD"    env-required:"true"`
	Addr     string `env:"POSTGRES_ADDR"        env-required:"true"`
	DB       string `env:"POSTGRES_DB"          env-required:"true"`
}

func MustLoad() *Config {
	cfg := &Config{}
	if err := cleanenv.ReadConfig(".env", cfg); err != nil {
		panic(err)
	}
	return cfg
}
