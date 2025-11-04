package admin

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/predicate"
	"famoria/internal/database/mongo/repositories/admin"

	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Bh        *th.BotHandler
	Log       *zap.Logger
	AdminRepo admin.Repository
	Cm        *callback.CallbacksManager
}

func Register(opts Opts) {
	opts.Bh.Handle(sendTextCmd{
		log: opts.Log,
	}.Handle, th.And(
		th.CommandEqual("text"),
		predicate.AdminCommand(opts.AdminRepo, 1),
	))
}
