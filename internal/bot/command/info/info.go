package info

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/database/mongo/repositories/brak"
)

type Opts struct {
	fx.In
	Bh       *th.BotHandler
	Log      *zap.Logger
	Cm       *callback.CallbacksManager
	BrakRepo brak.Repository
}

func Register(opts Opts) {
	opts.Bh.Handle(help{
		braks: opts.BrakRepo,
	}.Handle, th.And(
		th.Or(th.CommandEqual("help"), th.CommandEqual("start")),
	))

	opts.Bh.Handle(menu{
		braks: opts.BrakRepo,
	}.Handle, th.And(
		th.CommandEqual("menu"),
	))

	opts.Bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(&telego.SendMessageParams{
			ChatID:      tu.ID(update.Message.Chat.ID),
			Text:        "Меню закрыто, повторно открыть его можно написав /menu.",
			ReplyMarkup: tu.ReplyKeyboardRemove(),
		})
	}, th.And(
		th.Or(th.CommandEqual("closemenu"), th.TextEqual("❌ Закрыть")),
	))
}
