package family

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
)

func Register(bh *th.BotHandler, log *zap.Logger, cm *callback.CallbacksManager) {

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		data := cm.DynamicCallback("profile", callback.OneClick, []int64{update.Message.From.ID}, 5, "", func(query telego.CallbackQuery) {
			_, err := bot.SendMessage(tu.Messagef(
				telego.ChatID{ID: query.Message.GetChat().ID},
				"Hello %s!", query.From.FirstName,
			))
			if err != nil {
				log.Sugar().Error(err)
				return
			}
		})
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!", update.Message.From.FirstName,
		).WithReplyMarkup(
			tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					tu.InlineKeyboardButton("profile").
						WithCallbackData(data),
				),
			),
		))
	}, th.Or(th.CommandEqual("profile"), th.TextEqual("👤 Профиль")))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.CommandEqual("gobrak"))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.CommandEqual("endbrak"))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.Or(th.CommandEqual("braks"), th.TextEqual("💬 Браки чата")))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.Or(th.CommandEqual("braksglobal"), th.TextEqual("🌍 Браки всех чатов")))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.CommandEqual("kid"))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.CommandEqual("kidannihilate"))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.CommandEqual("detdom"))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.Or(th.CommandEqual("tree"), th.TextEqual("🌱 Древо (текст)")))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
		))
	}, th.Or(th.CommandEqual("treeimage"), th.TextEqual("🌳 Древо (картинка)")))
}
