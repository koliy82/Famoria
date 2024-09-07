package payments

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

func Register(bh *th.BotHandler, log *zap.Logger) {
	bh.HandlePreCheckoutQuery(func(bot *telego.Bot, update telego.PreCheckoutQuery) {
		log.Sugar().Info("PreCheckoutQuery: ", update)
		err := bot.AnswerPreCheckoutQuery(&telego.AnswerPreCheckoutQueryParams{
			PreCheckoutQueryID: update.ID,
			Ok:                 true,
		})
		if err != nil {
			log.Sugar().Error(err)
		}
	})

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Данная команда пока не реализована :(",
		))
	}, th.SuccessPayment())
}
