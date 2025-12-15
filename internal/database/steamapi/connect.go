package steamapi

import (
	"famoria/internal/config"
	"famoria/internal/database/steamapi/repositories/steam_accounts"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Log *zap.Logger
	Cfg config.Config
}

func New(opts Opts) *steam_accounts.SteamAPI {
	return &steam_accounts.SteamAPI{
		URL:    opts.Cfg.SteamAPIURI,
		ApiKey: opts.Cfg.SteamAPIKEY,
		Log:    opts.Log,
		Client: &http.Client{},
	}
}
