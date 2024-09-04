package app

import (
	"famoria/internal/bot"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/callback/static"
	"famoria/internal/bot/command/admin"
	"famoria/internal/bot/command/family"
	"famoria/internal/bot/command/info"
	"famoria/internal/bot/command/shop"
	"famoria/internal/bot/command/subscription"
	"famoria/internal/bot/handler"
	"famoria/internal/bot/handler/logger"
	"famoria/internal/bot/idle/item"
	shop2 "famoria/internal/bot/idle/item/shop"
	"famoria/internal/config"
	"famoria/internal/database/clickhouse"
	"famoria/internal/database/clickhouse/repositories/message"
	"famoria/internal/database/mongo"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"go.uber.org/fx"
)

var App = fx.Options(
	fx.Provide(
		config.New,
		config.SetupLogger,
	),
	fx.Provide(
		clickhouse.New,
		fx.Annotate(message.New, fx.As(new(message.Repository))),
		mongo.New,
		fx.Annotate(user.New, fx.As(new(user.Repository))),
		fx.Annotate(brak.New, fx.As(new(brak.Repository))),
		item.New,
		shop2.New,
	),
	fx.Provide(
		bot.New,
		handler.New,
		callback.New,
	),
	fx.Invoke(
		static.ProfileCallbacks,
	),
	fx.Invoke(
		admin.Register,
		family.Register,
		info.Register,
		shop.Register,
		subscription.Register,
		callback.Register,
		logger.Register,
		bot.PrintMe,
		handler.StartHandle,
	),
)
