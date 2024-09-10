package idle

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/config"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
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

	opts.Bh.Handle(shopCmd{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
		manager:  opts.M,
	}.Handle, th.Or(th.CommandEqual("shop"), th.TextEqual("🛒 Магазин")))

	opts.Bh.Handle(inventoryCmd{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
		manager:  opts.M,
	}.Handle, th.Or(th.CommandEqual("inventory"), th.TextEqual("🎒 Инвентарь")))
}
