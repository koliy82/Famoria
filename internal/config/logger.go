package config

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

type telegramCore struct {
	zapcore.Core
	bot    *telego.Bot
	chatID int64
}

func newTelegramCore(core zapcore.Core, bot *telego.Bot, chatID int64) zapcore.Core {
	return &telegramCore{core, bot, chatID}
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

	if c.ErrorsChatID != 0 {
		core, _ := config.Build()
		telegramCore := newTelegramCore(core.Core(), bot, c.ErrorsChatID)
		log = zap.New(telegramCore)
	} else {
		log, _ = config.Build()
	}

	zap.ReplaceGlobals(log)
	return log
}

func (t *telegramCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if ce := t.Core.Check(ent, ce); ce != nil {
		if ent.Level >= zapcore.ErrorLevel {
			go t.sendToTelegram(ent)
		}
		return ce
	}
	return ce
}

func (t *telegramCore) sendToTelegram(ent zapcore.Entry) {
	_, _ = t.bot.SendMessage(&telego.SendMessageParams{
		ChatID: tu.ID(t.chatID),
		Text: fmt.Sprintf("[%s]\n%s: %s", ent.Time.Format(time.RFC850),
			ent.Level.CapitalString(), ent.Message,
		),
	})
}
