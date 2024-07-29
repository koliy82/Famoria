package info

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go_tg_bot/internal/database/mongo/repositories/brak"
)

type menu struct {
	braks brak.Repository
}

func GenerateButtons(braks brak.Repository, userID int64) *telego.ReplyKeyboardMarkup {
	var rows [][]telego.KeyboardButton
	userBrak, _ := braks.FindByUserID(userID)
	if userBrak != nil {
		rows = append(rows, []telego.KeyboardButton{
			tu.KeyboardButton("👤 Профиль"),
			tu.KeyboardButton("💔 Развод"),
		})
	} else {
		rows = append(rows, []telego.KeyboardButton{
			tu.KeyboardButton("👤 Профиль"),
		})
	}

	kidBrak, _ := braks.FindByKidID(userID)
	if kidBrak != nil {
		if userBrak != nil && userBrak.BabyUserID != nil {
			rows = append(rows, []telego.KeyboardButton{
				tu.KeyboardButton("👶 Аннигиляция"),
				tu.KeyboardButton("🏠 Детдом"),
			})
		} else {
			rows = append(rows, []telego.KeyboardButton{
				tu.KeyboardButton("🏠 Детдом"),
			})
		}
	} else if userBrak != nil && userBrak.BabyUserID != nil {
		rows = append(rows, []telego.KeyboardButton{
			tu.KeyboardButton("👶 Аннигиляция"),
		})
	}
	rows = append(rows, tu.KeyboardRow(
		tu.KeyboardButton("🌱 Древо (текст)"),
		tu.KeyboardButton("🌳 Древо (картинка)"),
	))
	rows = append(rows, tu.KeyboardRow(
		tu.KeyboardButton("❌ Закрыть"),
	))
	return &telego.ReplyKeyboardMarkup{
		Keyboard:              rows,
		ResizeKeyboard:        true,
		InputFieldPlaceholder: "zxc",
		Selective:             true,
	}
}

func (m menu) Handle(bot *telego.Bot, update telego.Update) {
	//if update.Message.Chat.Type != "private" {
	//	_, _ = bot.SendMessage(&telego.SendMessageParams{
	//		ChatID: update.Message.Chat.ChatID(),
	//		Text:   "Меню работает только в личных сообщениях.",
	//	})
	//	return
	//}

	_, _ = bot.SendMessage(&telego.SendMessageParams{
		ChatID: tu.ID(update.Message.Chat.ID),
		Text:   "Меню показано ✅",
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.MessageID,
			AllowSendingWithoutReply: true,
		},
		ReplyMarkup: GenerateButtons(m.braks, update.Message.From.ID),
	})
}
