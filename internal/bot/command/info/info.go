package info

import (
	"context"

	"famoria/internal/bot/callback"
	"famoria/internal/bot/predicate"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/pkg/i18n"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Opts struct {
	fx.In
	Bh           *th.BotHandler
	Log          *zap.Logger
	Cm           *callback.CallbacksManager
	BrakRepo     brak.Repository
	ChatSettings chat_settings.Repository
}

func Register(opts Opts) {
	opts.Bh.Handle(helpCmd{
		brakRepo:     opts.BrakRepo,
		chatSettings: opts.ChatSettings,
		log:          opts.Log,
	}.Handle, th.And(
		th.Or(th.CommandEqual("help"), th.CommandEqual("start")),
	))

	opts.Bh.Handle(menuCmd{
		brakRepo:     opts.BrakRepo,
		chatSettings: opts.ChatSettings,
		log:          opts.Log,
	}.Handle, th.And(
		th.CommandEqual("menu"),
	))

	opts.Bh.Handle(func(ctx *th.Context, update telego.Update) error {
		lang := opts.ChatSettings.Lang(update.Message.Chat.ID)
		_, err := ctx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   i18n.T(lang, i18n.KeyMenuClosed),
			ReplyParameters: &telego.ReplyParameters{
				MessageID:                update.Message.GetMessageID(),
				AllowSendingWithoutReply: true,
			},
			ReplyMarkup: tu.ReplyKeyboardRemove().WithSelective(),
		})
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
		return err
	}, th.And(
		// The close button is localized, so both variants must be matched.
		th.Or(th.CommandEqual("closemenu"), predicate.TextEqualKey(i18n.KeyBtnClose)),
	))
}
