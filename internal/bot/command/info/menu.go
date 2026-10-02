package info

import (
	"context"

	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/pkg/i18n"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type menuCmd struct {
	brakRepo     brak.Repository
	chatSettings chat_settings.Repository
	log          *zap.Logger
}

// GenerateButtons builds the reply keyboard in the given language.
//
// Handlers match these labels by text, so the labels and their dispatch
// predicates must both go through i18n — see predicate.TextEqualKey.
func GenerateButtons(brakRepo brak.Repository, userID int64, lang i18n.Lang) *telego.ReplyKeyboardMarkup {
	var rows [][]telego.KeyboardButton
	userBrak, _ := brakRepo.FindByUserID(userID, nil)
	if userBrak != nil {
		rows = append(rows, []telego.KeyboardButton{
			tu.KeyboardButton(i18n.T(lang, i18n.KeyBtnProfile)),
			tu.KeyboardButton(i18n.T(lang, i18n.KeyBtnDivorce)),
		})
	} else {
		rows = append(rows, []telego.KeyboardButton{
			tu.KeyboardButton(i18n.T(lang, i18n.KeyBtnProfile)),
		})
	}

	kidBrak, _ := brakRepo.FindByKidID(userID)
	if kidBrak != nil {
		if userBrak != nil && userBrak.BabyUserID != nil {
			rows = append(rows, []telego.KeyboardButton{
				tu.KeyboardButton(i18n.T(lang, i18n.KeyBtnAnnihilate)),
				tu.KeyboardButton(i18n.T(lang, i18n.KeyBtnOrphanage)),
			})
		} else {
			rows = append(rows, []telego.KeyboardButton{
				tu.KeyboardButton(i18n.T(lang, i18n.KeyBtnOrphanage)),
			})
		}
	} else if userBrak != nil && userBrak.BabyUserID != nil {
		rows = append(rows, []telego.KeyboardButton{
			tu.KeyboardButton(i18n.T(lang, i18n.KeyBtnAnnihilate)),
		})
	}
	rows = append(rows, tu.KeyboardRow(
		tu.KeyboardButton(i18n.T(lang, i18n.KeyBtnTree)),
	))
	rows = append(rows, tu.KeyboardRow(
		tu.KeyboardButton(i18n.T(lang, i18n.KeyBtnClose)),
	))
	return &telego.ReplyKeyboardMarkup{
		Keyboard:              rows,
		ResizeKeyboard:        true,
		InputFieldPlaceholder: "zxc",
		Selective:             true,
	}
}

func (c menuCmd) Handle(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID
	lang := c.chatSettings.Lang(chatID)

	_, err := ctx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: tu.ID(chatID),
		Text:   i18n.T(lang, i18n.KeyMenuShown),
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.MessageID,
			AllowSendingWithoutReply: true,
		},
		ReplyMarkup: GenerateButtons(c.brakRepo, update.Message.From.ID, lang),
	})
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
