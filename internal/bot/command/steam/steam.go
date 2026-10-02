package steam

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/handler/waiter"
	"famoria/internal/bot/predicate"
	"famoria/internal/database/steamapi/repositories/steam_accounts"
	"famoria/internal/pkg/i18n"

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
	Mw  *waiter.MessageWaiter
}

func Register(opts Opts) {
	opts.Bh.Handle(listCmd{
		api: opts.Api,
		log: opts.Log,
		cm:  opts.Cm,
		mw:  opts.Mw,
	}.Handle, th.Or(th.CommandEqual("steam"), predicate.TextEqualKey(i18n.KeyBtnSteam)))
}
