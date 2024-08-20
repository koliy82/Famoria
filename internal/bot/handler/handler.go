package handler

import (
	"github.com/koliy82/telego"
	th "github.com/koliy82/telego/telegohandler"
)

func New(bot *telego.Bot) *th.BotHandler {
	var updates, _ = bot.UpdatesViaLongPolling(

		&telego.GetUpdatesParams{
			Timeout: 8,
			AllowedUpdates: []string{
				telego.MessageUpdates,
				telego.EditedMessageUpdates,
				telego.ChannelPostUpdates,
				telego.EditedChannelPostUpdates,
				//telego.MessageReaction,
				//telego.MessageReactionCount,
				telego.InlineQueryUpdates,
				telego.ChosenInlineResultUpdates,
				telego.CallbackQueryUpdates,
				telego.ShippingQueryUpdates,
				telego.PreCheckoutQueryUpdates,
				telego.PollUpdates,
				telego.PollAnswerUpdates,
				telego.MyChatMemberUpdates,
				telego.ChatMemberUpdates,
				telego.ChatJoinRequestUpdates,
			},
		},
	)

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
