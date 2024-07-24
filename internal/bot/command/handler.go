package command

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"
	"os"
)

func New(bot *telego.Bot, log *zap.Logger) *th.BotHandler {
	updates, _ := bot.UpdatesViaLongPolling(nil)

	bh, err := th.NewBotHandler(bot, updates)

	if err != nil {
		log.Sugar().Error(err)
		os.Exit(1)
	}

	return bh
}

func StartHandle(bot *telego.Bot, bh *th.BotHandler) {

	defer bh.Stop()

	defer bot.StopLongPolling()

	bh.Start()
}
