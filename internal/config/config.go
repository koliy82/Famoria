package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"famoria/internal/pkg/proxy"

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

	// BotProxy routes all Telegram Bot API traffic through BotProxyURL when true.
	// Needed where api.telegram.org is blocked at the host level.
	BotProxy bool `envconfig:"BOT_PROXY" default:"false"`

	// BotProxyURL is the HTTP(S) CONNECT proxy for the Telegram API, in
	// "user:pass@host:port" form. Required when BotProxy is true.
	BotProxyURL *string `envconfig:"BOT_PROXY_URL"`

	// DBProxy routes all MongoDB traffic through DBProxyURL when true. Needed
	// where the database host is blocked at the host level.
	DBProxy bool `envconfig:"DB_PROXY" default:"false"`

	// DBProxyURL is the HTTP(S) CONNECT proxy for MongoDB, in
	// "user:pass@host:port" form. Required when DBProxy is true.
	//
	// MongoDB's wire protocol is not HTTP, so this is a raw CONNECT tunnel rather
	// than an HTTP proxy setting; the driver still negotiates its own TLS inside
	// the tunnel.
	DBProxyURL *string `envconfig:"DB_PROXY_URL"`
}

// lowercaseEnvKeys are names that may appear in .env in lower case.
//
// envconfig only ever looks up upper-case names, and godotenv preserves the
// spelling exactly as written in .env. On Windows environment variables are
// case-insensitive so either spelling works, but in Linux containers they are
// case-sensitive: a lower-case "bot_proxy=true" would be silently ignored and
// the bot would connect direct. Promoting the known keys here makes both
// spellings work everywhere.
var lowercaseEnvKeys = []string{
	"bot_proxy", "bot_proxy_url",
	"db_proxy", "db_proxy_url",
}

// normalizeEnvKeys promotes the lower-case spellings of known keys to the
// upper-case form envconfig looks up, without overwriting a value that was
// already set explicitly.
func normalizeEnvKeys() {
	promoteKeys(lowercaseEnvKeys, os.LookupEnv, func(key, value string) {
		_ = os.Setenv(key, value)
	})
}

// promoteKeys copies each lower-case key's value to its upper-case name when the
// upper-case name is unset.
//
// The lookup and set functions are injected rather than calling os directly:
// environment variables are case-insensitive on Windows but case-sensitive on
// Linux, so testing this through the real environment would pass trivially on one
// platform and fail on the other regardless of whether the logic is correct.
func promoteKeys(keys []string, lookup func(string) (string, bool), set func(key, value string)) {
	for _, key := range keys {
		upper := strings.ToUpper(key)
		if _, exists := lookup(upper); exists {
			continue
		}
		if v, ok := lookup(key); ok {
			set(upper, v)
		}
	}
}

func New() Config {
	cfg := Config{}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	envPath := filepath.Join(wd, ".env")

	_ = godotenv.Load(envPath)
	normalizeEnvKeys()
	if err := envconfig.Process("", &cfg); err != nil {
		panic(err)
	}

	loc, err := time.LoadLocation(cfg.AppTimeZone)
	if err != nil {
		panic(err)
	}
	time.Local = loc

	if err := cfg.validateProxies(); err != nil {
		panic(err)
	}

	return cfg
}

// validateProxies rejects a half-configured proxy.
//
// Without this, a missing or malformed URL would fall back to a direct
// connection, which on a network where the destination is blocked surfaces as an
// opaque timeout rather than a clear startup error.
func (c Config) validateProxies() error {
	for _, p := range []struct {
		name    string
		enabled bool
		url     *string
	}{
		{"BOT_PROXY", c.BotProxy, c.BotProxyURL},
		{"DB_PROXY", c.DBProxy, c.DBProxyURL},
	} {
		if !p.enabled {
			continue
		}
		if p.url == nil || strings.TrimSpace(*p.url) == "" {
			return fmt.Errorf("config: %s is true but %s_URL is empty", p.name, p.name)
		}
		if _, err := proxy.Parse(*p.url); err != nil {
			return fmt.Errorf("config: invalid %s_URL: %w", p.name, err)
		}
	}
	return nil
}
