package info

import (
	"context"
	"famoria/internal/database/mongo/repositories/brak"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type menuCmd struct {
	brakRepo brak.Repository
	log      *zap.Logger
}

func GenerateButtons(brakRepo brak.Repository, userID int64) *telego.ReplyKeyboardMarkup {
	var rows [][]telego.KeyboardButton
	userBrak, _ := brakRepo.FindByUserID(userID, nil)
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

	kidBrak, _ := brakRepo.FindByKidID(userID)
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
		tu.KeyboardButton("🌱 Семейное древо"),
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

func (c menuCmd) Handle(ctx *th.Context, update telego.Update) error {
	_, err := ctx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: tu.ID(update.Message.Chat.ID),
		Text:   "Меню показано ✅",
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.MessageID,
			AllowSendingWithoutReply: true,
		},
		ReplyMarkup: GenerateButtons(c.brakRepo, update.Message.From.ID),
	})
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
