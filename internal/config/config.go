package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	AppEnv        string `envconfig:"APP_ENV" default:"dev"`
	AppTimeZone   string `envconfig:"APP_TIMEZONE" default:"Europe/Moscow"`
	TelegramToken string `envconfig:"TELEGRAM_TOKEN" required:"true"`
	ErrorsChatID  int64  `envconfig:"ERRORS_CHAT_ID" default:"0"`

	ClickhouseURL      string `envconfig:"CLICKHOUSE_URL" required:"true"`
	ClickhousePort     int    `envconfig:"CLICKHOUSE_PORT" required:"true"`
	ClickhouseUser     string `envconfig:"CLICKHOUSE_USER" required:"true"`
	ClickhousePassword string `envconfig:"CLICKHOUSE_PASSWORD" required:"true"`
	ClickhouseDatabase string `envconfig:"CLICKHOUSE_DATABASE" default:"koliy82"`

	MongoURI      string `envconfig:"MONGO_URI" required:"true"`
	MongoDatabase string `envconfig:"MONGO_DATABASE" required:"true"`
}

func New() Config {
	cfg := Config{}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	envPath := filepath.Join(wd, ".env")

	_ = godotenv.Load(envPath)

	if err := envconfig.Process("", &cfg); err != nil {
		panic(err)
	}

	loc, err := time.LoadLocation(cfg.AppTimeZone)
	if err != nil {
		panic(err)
	}
	time.Local = loc

	return cfg
}
