package logger

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go_tg_bot/internal/database/clickhouse/repositories/message"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
)

type MessageLogger struct {
	messages message.Repository
	users    user.Repository
	braks    brak.Repository
}

func (l MessageLogger) Handle(bot *telego.Bot, update telego.Update) {
	msg := update.Message
	from := msg.From
	if from == nil {
		return
	}
	err := l.users.ValidateInfo(update.Message.From)
	if err != nil {
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

type Opts struct {
	fx.In
	Bh          *th.BotHandler
	Log         *zap.Logger
	MessageRepo message.Repository
	UserRepo    user.Repository
	BrakRepo    brak.Repository
}

func Register(opts Opts) {
	opts.Bh.Handle(MessageLogger{
		messages: opts.MessageRepo,
		users:    opts.UserRepo,
		braks:    opts.BrakRepo,
	}.Handle, th.AnyMessage())
}
