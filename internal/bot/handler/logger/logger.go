package logger

import (
	"famoria/internal/bot/handler/waiter"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/message"
	"famoria/internal/database/mongo/repositories/user"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MessageLogger struct {
	messageRepo message.Repository
	userRepo    user.Repository
	mw          *waiter.MessageWaiter
	log         *zap.Logger
}

func (l MessageLogger) Handle(ctx *th.Context, update telego.Update) error {
	l.mw.HandleMessageUpdate(ctx, update)
	msg := update.Message
	from := msg.From
	if from == nil {
		return nil
	}
	_, err := l.userRepo.FindOrUpdate(update.Message.From)
	if err != nil {
		return err
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
	return nil
}

type Opts struct {
	fx.In
	Bh          *th.BotHandler
	Log         *zap.Logger
	MessageRepo message.Repository
	UserRepo    user.Repository
	BrakRepo    brak.Repository
	Mw          *waiter.MessageWaiter
}

func Register(opts Opts) {
	opts.Bh.Handle(MessageLogger{
		messageRepo: opts.MessageRepo,
		userRepo:    opts.UserRepo,
		mw:          opts.Mw,
		log:         opts.Log,
	}.Handle, th.AnyMessage())
}
