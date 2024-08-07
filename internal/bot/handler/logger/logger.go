package logger

import (
	"famoria/internal/database/clickhouse/repositories/message"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MessageLogger struct {
	messageRepo message.Repository
	userRepo    user.Repository
	brakRepo    brak.Repository
}

func (l MessageLogger) Handle(bot *telego.Bot, update telego.Update) {
	msg := update.Message
	from := msg.From
	if from == nil {
		return
	}
	_, err := l.userRepo.FindOrUpdate(update.Message.From)
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

	l.messageRepo.Insert(
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
		messageRepo: opts.MessageRepo,
		userRepo:    opts.UserRepo,
		brakRepo:    opts.BrakRepo,
	}.Handle, th.AnyMessage())
}
