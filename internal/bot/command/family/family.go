package family

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/database/clickhouse/repositories/message"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
)

type Opts struct {
	fx.In
	Bh          *th.BotHandler
	Log         *zap.Logger
	BrakRepo    brak.Repository
	UserRepo    user.Repository
	MessageRepo message.Repository
	Cm          *callback.CallbacksManager
}

func Register(opts Opts) {

	opts.Bh.Handle(profile{
		cm:          opts.Cm,
		brakRepo:    opts.BrakRepo,
		userRepo:    opts.UserRepo,
		messageRepo: opts.MessageRepo,
		log:         opts.Log,
	}.Handle, th.Or(th.CommandEqual("profile"), th.TextEqual("👤 Профиль"), th.CommandEqual("mybrak")))

	opts.Bh.Handle(goFamily{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		log:      opts.Log,
	}.Handle, th.CommandEqual("gobrak"))

	opts.Bh.Handle(endFamily{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
	}.Handle, th.Or(th.CommandEqual("endbrak"), th.TextEqual("💔 Развод")))

	opts.Bh.Handle(goKid{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
	}.Handle, th.CommandEqual("kid"))

	opts.Bh.Handle(endKid{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
	}.Handle, th.Or(th.CommandEqual("kidannihilate"), th.TextEqual("👶 Аннигиляция")))

	opts.Bh.Handle(leaveKid{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		userRepo: opts.UserRepo,
		log:      opts.Log,
	}.Handle, th.Or(th.CommandEqual("detdom"), th.TextEqual("🏠 Детдом")))

	opts.Bh.Handle(brakPages{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		isLocal:  true,
		log:      opts.Log,
	}.Handle, th.Or(th.CommandEqual("braks"), th.TextEqual("💬 Браки чата")))

	opts.Bh.Handle(brakPages{
		cm:       opts.Cm,
		brakRepo: opts.BrakRepo,
		isLocal:  false,
		log:      opts.Log,
	}.Handle, th.Or(th.CommandEqual("braksglobal"), th.TextEqual("🌍 Браки всех чатов")))

	opts.Bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.Or(th.CommandEqual("tree"), th.TextEqual("🌱 Древо (текст)")))

	opts.Bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.Or(th.CommandEqual("treeimage"), th.TextEqual("🌳 Древо (картинка)")))
}
