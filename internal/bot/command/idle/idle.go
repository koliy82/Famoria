package idle

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/predicate"
	"famoria/internal/config"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/i18n"

	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Bh           *th.BotHandler
	Log          *zap.Logger
	Cfg          config.Config
	BrakRepo     brak.Repository
	UserRepo     user.Repository
	ChatSettings chat_settings.Repository
	Cm           *callback.CallbacksManager
	M            *item.Manager
}

func Register(opts Opts) {

	opts.Bh.Handle(shopCmd{
		cm:           opts.Cm,
		brakRepo:     opts.BrakRepo,
		userRepo:     opts.UserRepo,
		chatSettings: opts.ChatSettings,
		log:          opts.Log,
		manager:      opts.M,
	}.Handle, th.And(
		th.Or(th.CommandEqual("shop"), predicate.TextEqualKey(i18n.KeyBtnShop)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatShop),
	))

	opts.Bh.Handle(inventoryCmd{
		cm:           opts.Cm,
		brakRepo:     opts.BrakRepo,
		userRepo:     opts.UserRepo,
		chatSettings: opts.ChatSettings,
		log:          opts.Log,
		manager:      opts.M,
	}.Handle, th.And(
		th.Or(th.CommandEqual("inventory"), predicate.TextEqualKey(i18n.KeyBtnInventory)),
		predicate.ChatFeatureEnabled(opts.ChatSettings, chat_settings.FeatInventory),
	))
}
