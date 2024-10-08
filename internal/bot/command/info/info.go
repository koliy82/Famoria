package info

import (
	"famoria/internal/bot/callback"
	"famoria/internal/database/mongo/repositories/brak"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Bh       *th.BotHandler
	Log      *zap.Logger
	Cm       *callback.CallbacksManager
	BrakRepo brak.Repository
}

func Register(opts Opts) {
	opts.Bh.Handle(helpCmd{
		brakRepo: opts.BrakRepo,
		log:      opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqual("help"), th.CommandEqual("start")),
	))

	opts.Bh.Handle(menuCmd{
		brakRepo: opts.BrakRepo,
		log:      opts.Log,
	}.Handle, th.And(
		th.CommandEqual("menu"),
	))

	opts.Bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   "Меню закрыто, повторно открыть его можно написав /menu.",
			ReplyParameters: &telego.ReplyParameters{
				MessageID:                update.Message.GetMessageID(),
				AllowSendingWithoutReply: true,
			},
			ReplyMarkup: tu.ReplyKeyboardRemove().WithSelective(),
		})
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
	}, th.And(
		th.Or(th.CommandEqual("closemenu"), th.TextEqual("❌ Закрыть")),
	))
}
