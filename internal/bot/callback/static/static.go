package static

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
)

func Register(cm *callback.CallbacksManager, log *zap.Logger, bot *telego.Bot) {
	cm.StaticCallback("num", func(query telego.CallbackQuery) {

	})

	cm.StaticCallback("back", func(query telego.CallbackQuery) {
		Chat := query.Message.GetChat()
		MessageID := query.Message.GetMessageID()
		ChatID := Chat.ChatID()
		// TODO
		// find page id
		// paginate list and change text
		text, err := bot.EditMessageText(&telego.EditMessageTextParams{
			MessageID: MessageID,
			ChatID:    ChatID,
			Text:      "",
			ReplyMarkup: tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					tu.InlineKeyboardButton("⬅️").
						WithCallbackData("back"),
					tu.InlineKeyboardButton("id").
						WithCallbackData("num"),
					tu.InlineKeyboardButton("➡️").
						WithCallbackData("next"),
				),
			),
		})

		if err != nil {
			log.Sugar().Error(err)
			return
		}
		log.Sugar().Info(text)
	})

	cm.StaticCallback("next", func(query telego.CallbackQuery) {
		Chat := query.Message.GetChat()
		MessageID := query.Message.GetMessageID()
		ChatID := Chat.ChatID()
		// TODO
		// find page id
		// paginate list and change text
		text, err := bot.EditMessageText(&telego.EditMessageTextParams{
			MessageID: MessageID,
			ChatID:    ChatID,
			Text:      "",
			ReplyMarkup: tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					tu.InlineKeyboardButton("⬅️").
						WithCallbackData("back"),
					tu.InlineKeyboardButton("id").
						WithCallbackData("num"),
					tu.InlineKeyboardButton("➡️").
						WithCallbackData("next"),
				),
			),
		})

		if err != nil {
			log.Sugar().Error(err)
			return
		}
		log.Sugar().Info(text)
	})
}
