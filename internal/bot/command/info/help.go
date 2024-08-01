package info

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"strings"
)

type help struct {
	brakRepo brak.Repository
}

func (h help) Handle(bot *telego.Bot, update telego.Update) {
	commands, err := bot.GetMyCommands(&telego.GetMyCommandsParams{})
	if err != nil {
		return
	}
	text := ""
	for _, command := range commands {
		text += "/" + command.Command + " - " + command.Description + "\n"
	}
	_, _ = bot.SendMessage(&telego.SendMessageParams{
		ChatID: tu.ID(update.Message.Chat.ID),
		Text:   strings.TrimSpace(text),
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.MessageID,
			ChatID:                   tu.ID(update.Message.Chat.ID),
			AllowSendingWithoutReply: true,
		},
		ReplyMarkup: GenerateButtons(h.brakRepo, update.Message.From.ID),
	})
}
