package family

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
)

type profile struct {
	cm  *callback.CallbacksManager
	log *zap.Logger
}

func (p profile) Handle(bot *telego.Bot, update telego.Update) {
	callbackOne := p.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "one",
		CtxType:  callback.ChooseOne,
		OwnerIDs: []int64{update.Message.From.ID},
		Time:     5,
		Callback: func(query telego.CallbackQuery) {
			_, err := bot.SendMessage(tu.Messagef(
				telego.ChatID{ID: query.Message.GetChat().ID},
				"Hello %s!", query.From.FirstName,
			))
			if err != nil {
				p.log.Sugar().Error(err)
				return
			}
		},
	})

	callbackTwo := p.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "two",
		CtxType:  callback.ChooseOne,
		OwnerIDs: []int64{update.Message.From.ID},
		Time:     5,
		Callback: func(query telego.CallbackQuery) {
			_, err := bot.SendMessage(tu.Messagef(
				telego.ChatID{ID: query.Message.GetChat().ID},
				"Hello %s!", query.From.FirstName,
			))
			if err != nil {
				p.log.Sugar().Error(err)
				return
			}
		},
	})

	_, _ = bot.SendMessage(tu.Messagef(
		tu.ID(update.Message.Chat.ID),
		"Hello %s!", update.Message.From.FirstName,
	).WithReplyMarkup(
		tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				callbackOne.Inline(),
				callbackTwo.Inline(),
			),
		),
	))
}
