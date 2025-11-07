package app

import (
	"famoria/internal/bot"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/callback/static"
	"famoria/internal/bot/command/admin"
	"famoria/internal/bot/command/donate"
	"famoria/internal/bot/command/family"
	"famoria/internal/bot/command/idle"
	"famoria/internal/bot/command/info"
	"famoria/internal/bot/handler"
	"famoria/internal/bot/handler/logger"
	"famoria/internal/bot/handler/payments"
	"famoria/internal/bot/idle/item"
	"famoria/internal/config"
	"famoria/internal/database/clickhouse"
	"famoria/internal/database/mongo"
	admin2 "famoria/internal/database/mongo/repositories/admin"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/checkout"
	"famoria/internal/database/mongo/repositories/message"
	"famoria/internal/database/mongo/repositories/payment"
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
		fx.Annotate(admin2.New, fx.As(new(admin2.Repository))),
		fx.Annotate(checkout.New, fx.As(new(checkout.Repository))),
		fx.Annotate(payment.New, fx.As(new(payment.Repository))),
		item.New,
	),
	fx.Provide(
		bot.New,
		handler.New,
		callback.New,
	),
	fx.Invoke(
		static.ProfileCallbacks,
		payments.Register,
		admin.Register,
		family.Register,
		idle.Register,
		info.Register,
		donate.Register,
		callback.Register,
		logger.Register,
		bot.PrintMe,
		handler.StartHandle,
	),
)
