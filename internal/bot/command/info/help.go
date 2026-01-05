package info

import (
	"context"
	"famoria/internal/database/mongo/repositories/brak"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type helpCmd struct {
	brakRepo brak.Repository
	log      *zap.Logger
}

func (c helpCmd) Handle(ctx *th.Context, update telego.Update) error {
	commands, err := ctx.Bot().GetMyCommands(context.Background(), &telego.GetMyCommandsParams{})
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}
	text := "Основная концепция бота заключается в создании семей между пользователями поэтому основной функционал бота становится доступен после вступления в брак с другим пользователем.\n\nДоступные команды:\n"
	for _, command := range commands {
		text += "/" + command.Command + " - " + command.Description + "\n"
	}
	_, err = ctx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: tu.ID(update.Message.Chat.ID),
		Text:   strings.TrimSpace(text),
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.MessageID,
			ChatID:                   tu.ID(update.Message.Chat.ID),
			AllowSendingWithoutReply: true,
		},
		ReplyMarkup: GenerateButtons(c.brakRepo, update.Message.From.ID),
	})
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
