package app

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/bot/callback/static"
	"go_tg_bot/internal/bot/command"
	"go_tg_bot/internal/bot/command/admin"
	"go_tg_bot/internal/bot/command/family"
	"go_tg_bot/internal/bot/command/minecraft"
	"go_tg_bot/internal/bot/logger"
	"go_tg_bot/internal/config"
	"go_tg_bot/internal/database/clickhouse"
	"go_tg_bot/internal/database/clickhouse/repositories/message"
	"go_tg_bot/internal/database/mongo"
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
		mongo.New,
		clickhouse.New,
		fx.Annotate(user.New, fx.As(new(user.Repository))),
		fx.Annotate(message.New, fx.As(new(message.Repository))),
		//fx.Annotate(brak.New, fx.As(new(brak.Repository))),
	),
	fx.Provide(
		bot.New,
		command.New,
		callback.New,
	),
	fx.Invoke(
		admin.Register,
		family.Register,
		minecraft.Register,
		callback.Register,
		static.Register,
		logger.Register,
		command.StartHandle,
	),
)
