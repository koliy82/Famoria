package admin

import (
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/bot/predicate"
)

type Opts struct {
	fx.In
	Bh  *th.BotHandler
	Log *zap.Logger
	Cm  *callback.CallbacksManager
}

func Register(opts Opts) {
	opts.Bh.Handle(sendText{
		log: opts.Log,
	}.Handle, th.And(
		th.CommandEqual("text"),
		predicate.AdminCommand(),
	))
}
