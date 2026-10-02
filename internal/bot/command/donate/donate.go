package donate

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/predicate"
	"famoria/internal/config"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/i18n"

	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Bh       *th.BotHandler
	Log      *zap.Logger
	Cfg      config.Config
	BrakRepo brak.Repository
	UserRepo user.Repository
	Cm       *callback.CallbacksManager
	M        *item.Manager
}

func Register(opts Opts) {
	opts.Bh.Handle(SubscribeCmd{
		userRepo:    opts.UserRepo,
		brakRepo:    opts.BrakRepo,
		log:         opts.Log,
		cm:          opts.Cm,
		m:           opts.M,
		yKassaToken: opts.Cfg.YKassaToken,
	}.Handle, th.Or(th.CommandEqual("subscribe"), predicate.TextEqualKey(i18n.KeyBtnSubscribe)))
}
