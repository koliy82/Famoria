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

	// YtdlpCookiesFile is an optional path to a Netscape-format cookies file
	// used by yt-dlp. Required for YouTube, which now blocks unauthenticated
	// (bot) access. Only applied to YouTube URLs.
	YtdlpCookiesFile *string `envconfig:"YTDLP_COOKIES_FILE"`

	// ProxyEnable controls when ProxyURL is applied to yt-dlp requests:
	//   "false"   (default) — proxy is not used at all.
	//   "youtube"           — proxy is used only for YouTube URLs.
	//   "true"              — proxy is used for all URLs.
	ProxyEnable string `envconfig:"PROXY_ENABLE" default:"false"`

	// ProxyURL is an optional HTTP/HTTPS/SOCKS proxy URL used by yt-dlp, when
	// enabled by ProxyEnable. Useful when the server's datacenter IP is blocked
	// by YouTube ("Sign in to confirm you're not a bot"). A residential proxy is
	// typically required. Format: http://user:pass@host:port
	ProxyURL *string `envconfig:"PROXY_URL"`
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
