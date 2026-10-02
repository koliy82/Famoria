package info

import (
	"context"
	"strings"

	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/pkg/i18n"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type helpCmd struct {
	brakRepo     brak.Repository
	chatSettings chat_settings.Repository
	log          *zap.Logger
}

func (c helpCmd) Handle(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID
	lang := c.chatSettings.Lang(chatID)

	commands, err := ctx.Bot().GetMyCommands(context.Background(), &telego.GetMyCommandsParams{})
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}
	text := i18n.T(lang, i18n.KeyHelpIntro) + "\n"
	for _, command := range commands {
		text += i18n.T(lang, i18n.KeyHelpCommands, command.Command, command.Description) + "\n"
	}
	_, err = ctx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: tu.ID(chatID),
		Text:   strings.TrimSpace(text),
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.MessageID,
			ChatID:                   tu.ID(chatID),
			AllowSendingWithoutReply: true,
		},
		ReplyMarkup: GenerateButtons(c.brakRepo, update.Message.From.ID, lang),
	})
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
