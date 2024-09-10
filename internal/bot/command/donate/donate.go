package donate

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/config"
	"famoria/internal/database/clickhouse/repositories/message"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Bh          *th.BotHandler
	Log         *zap.Logger
	Cfg         config.Config
	BrakRepo    brak.Repository
	UserRepo    user.Repository
	MessageRepo message.Repository
	Cm          *callback.CallbacksManager
	M           *item.Manager
}

func Register(opts Opts) {
	opts.Bh.Handle(SubscribeCmd{
		userRepo: opts.UserRepo,
		brakRepo: opts.BrakRepo,
		log:      opts.Log,
		cm:       opts.Cm,
		m:        opts.M,
	}.Handle, th.Or(th.CommandEqual("subscribe"), th.TextEqual("💳 Подписка")))
}
