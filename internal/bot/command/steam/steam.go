package steam

import (
	"famoria/internal/bot/callback"
	"famoria/internal/database/steamapi/repositories/steam_accounts"

	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Bh  *th.BotHandler
	Log *zap.Logger
	Api *steam_accounts.SteamAPI
	Cm  *callback.CallbacksManager
}

func Register(opts Opts) {
	opts.Bh.Handle(listCmd{
		api: opts.Api,
		log: opts.Log,
		cm:  opts.Cm,
	}.Handle, th.Or(th.CommandEqual("steam"), th.TextEqual("🎮 Steam аккаунты")))
}
