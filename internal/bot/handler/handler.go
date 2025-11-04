package handler

import (
	"context"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func New(bot *telego.Bot) *th.BotHandler {
	updates, _ := bot.UpdatesViaLongPolling(context.Background(), nil)

	bh, err := th.NewBotHandler(bot, updates)

	if err != nil {
		panic(err)
	}

	return bh
}

func StartHandle(bot *telego.Bot, bh *th.BotHandler) {

	defer func() { _ = bh.Stop() }()

	//defer bot.StopLongPolling()

	_ = bh.Start()
}
