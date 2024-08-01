package handler

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func New(bot *telego.Bot) *th.BotHandler {
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
