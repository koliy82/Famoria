package admin

import (
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/bot/predicate"
	"go_tg_bot/internal/database/clickhouse/repositories/message"
	"go_tg_bot/internal/database/mongo/repositories/user"
)

type Opts struct {
	fx.In
	Bh          *th.BotHandler
	Log         *zap.Logger
	Cm          *callback.CallbacksManager
	MessageRepo message.Repository
	UserRepo    user.Repository
}

func Register(opts Opts) {
	opts.Bh.Handle(sendText{
		log: opts.Log,
	}.Handle, th.And(
		th.CommandEqual("text"),
		predicate.AdminCommand(),
	))

	opts.Bh.Handle(messageLogger{
		messages: opts.MessageRepo,
		users:    opts.UserRepo,
	}.Handle, th.AnyMessage())

}
