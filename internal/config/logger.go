package config

import (
	"context"
	"fmt"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type telegramCore struct {
	zapcore.Core
	bot         *telego.Bot
	infoChatID  *int64
	warnChatID  *int64
	errorChatID *int64
}

func (t *telegramCore) toChatID(level zapcore.Level) *int64 {
	switch level {
	case zapcore.InfoLevel:
		return t.infoChatID
	case zapcore.WarnLevel:
		return t.warnChatID
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return t.errorChatID
	default:
		return nil
	}
}

func newTelegramCore(core zapcore.Core, bot *telego.Bot, c Config) zapcore.Core {
	return &telegramCore{core, bot, c.InfoChatID, c.WarnChatID, c.ErrorsChatID}
}

func SetupLogger(c Config, bot *telego.Bot) *zap.Logger {
	var log *zap.Logger
	var config zap.Config

	switch c.AppEnv {
	case "prod":
		config = zap.NewProductionConfig()
	default:
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC850)

	if c.ErrorsChatID != nil {
		core, _ := config.Build()
		telegramCore := newTelegramCore(core.Core(), bot, c)
		log = zap.New(telegramCore)
	} else {
		log, _ = config.Build()
	}

	zap.ReplaceGlobals(log)
	return log
}

func (t *telegramCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if ce := t.Core.Check(ent, ce); ce != nil {
		if ent.Level >= zapcore.InfoLevel {
			go t.sendToTelegram(ent)
		}
		return ce
	}
	return ce
}

func (t *telegramCore) sendToTelegram(ent zapcore.Entry) {
	chatId := t.toChatID(ent.Level)
	if chatId == nil {
		return
	}
	_, _ = t.bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: tu.ID(*chatId),
		Text: fmt.Sprintf("[%s]\n%s: %s", ent.Time.Format(time.RFC850),
			ent.Level.CapitalString(), ent.Message,
		),
	})

}
