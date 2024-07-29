package app

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/bot/callback/static"
	"go_tg_bot/internal/bot/command/admin"
	"go_tg_bot/internal/bot/command/family"
	"go_tg_bot/internal/bot/command/info"
	"go_tg_bot/internal/bot/command/minecraft"
	"go_tg_bot/internal/bot/handler"
	"go_tg_bot/internal/bot/handler/logger"
	"go_tg_bot/internal/config"
	"go_tg_bot/internal/database/clickhouse"
	"go_tg_bot/internal/database/clickhouse/repositories/message"
	"go_tg_bot/internal/database/mongo"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
)

var App = fx.Options(
	fx.Provide(
		func() *zap.Logger {
			log, _ := zap.NewDevelopment()
			zap.ReplaceGlobals(log)
			return log
		},
		config.New,
	),
	fx.Provide(
		clickhouse.New,
		fx.Annotate(message.New, fx.As(new(message.Repository))),
		mongo.New,
		fx.Annotate(user.New, fx.As(new(user.Repository))),
		fx.Annotate(brak.New, fx.As(new(brak.Repository))),
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
		info.Register,
		admin.Register,
		family.Register,
		minecraft.Register,
		callback.Register,
		static.Register,
		logger.Register,
		handler.StartHandle,
	),
)
