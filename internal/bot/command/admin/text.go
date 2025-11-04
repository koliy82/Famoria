package admin

import (
	"context"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type sendTextCmd struct {
	log *zap.Logger
}

func (c sendTextCmd) Handle(ctx *th.Context, update telego.Update) error {
	chatID := tu.ID(update.Message.Chat.ID)
	args := strings.Split(update.Message.Text, " ")
	err := ctx.Bot().DeleteMessage(
		context.Background(),
		&telego.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: update.Message.MessageID,
		},
	)
	if err != nil {
		c.log.Error(err.Error())
		return err
	}
	_, err = ctx.Bot().SendMessage(
		context.Background(),
		tu.Messagef(
			chatID,
			strings.Join(args[1:], " "),
		),
	)
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}
	return nil
}
