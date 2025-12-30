package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppEnv            string  `envconfig:"APP_ENV" default:"dev"`
	AppTimeZone       string  `envconfig:"APP_TIMEZONE" default:"Europe/Moscow"`
	TelegramToken     string  `envconfig:"TELEGRAM_TOKEN" required:"true"`
	TelegramTestToken *string `envconfig:"TELEGRAM_TEST_TOKEN"`

	InfoChatID   *int64 `envconfig:"INFO_CHAT_ID"`
	WarnChatID   *int64 `envconfig:"WARN_CHAT_ID"`
	ErrorsChatID *int64 `envconfig:"ERRORS_CHAT_ID"`

	MongoURI              string  `envconfig:"MONGO_URI" required:"true"`
	MongoDatabase         string  `envconfig:"MONGO_DATABASE" required:"true"`
	TransferMongoDatabase *string `envconfig:"TRANSFER_MONGO_DATABASE"`
	MongoSteamDatabase    *string `envconfig:"MONGO_STEAM_DATABASE"`
	MongoFarmLogsCollName *string `envconfig:"MONGO_FARM_LOGS_COLL_NAME"`

	TreeApiURL string `envconfig:"TREE_API_URL" default:"http://localhost:8000"`

	YKassaToken *string `envconfig:"YKASSA_TOKEN" required:"false"`

	SteamURI string `envconfig:"STEAM_URL"`
	SteamKEY string `envconfig:"STEAM_KEY"`
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
