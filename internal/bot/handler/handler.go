package handler

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"
)

func New(bot *telego.Bot, log *zap.Logger) *th.BotHandler {
	updates, _ := bot.UpdatesViaLongPolling(nil)

	bh, err := th.NewBotHandler(bot, updates)

	if err != nil {
		panic(err)
	}

	return bh
}

func StartHandle(bot *telego.Bot, bh *th.BotHandler) {

	defer bh.Stop()

	defer bot.StopLongPolling()

	bh.Start()
}
