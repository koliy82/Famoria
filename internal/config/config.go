package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"os"
	"path/filepath"
)

type Config struct {
	AppEnv string `envconfig:"APP_ENV" default:"dev"`

	TelegramToken string `envconfig:"TELEGRAM_TOKEN" required:"true"`

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

	return cfg
}
