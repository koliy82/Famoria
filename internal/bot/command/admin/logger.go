package admin

import (
	"github.com/mymmrac/telego"
	"go_tg_bot/internal/database/clickhouse/repositories/message"
	"go_tg_bot/internal/database/mongo/repositories/user"
)

type messageLogger struct {
	messages message.Repository
	users    user.Repository
}

func (l messageLogger) Handle(bot *telego.Bot, update telego.Update) {
	msg := update.Message
	from := msg.From
	if from == nil {
		return
	}

	newMessage := &message.Message{
		ID:      msg.MessageID,
		ChatID:  msg.Chat.ID,
		UserID:  msg.From.ID,
		Date:    msg.Date,
		Text:    &msg.Text,
		Caption: &msg.Caption,
	}

	if msg.ReplyToMessage != nil {
		newMessage.ReplyID = &msg.ReplyToMessage.MessageID
	}

	l.messages.Insert(
		newMessage,
	)
}
