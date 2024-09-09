package idle

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/shop"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type shopCmd struct {
	brakRepo brak.Repository
	userRepo user.Repository
	cm       *callback.CallbacksManager
	log      *zap.Logger
	manager  *item.Manager
}

func (c shopCmd) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	b, _ := c.brakRepo.FindByUserID(from.ID, c.manager)
	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.GetMessageID(),
			AllowSendingWithoutReply: true,
		},
	}
	if b == nil {
		_, err := bot.SendMessage(params.WithText("Для просмотра инвентаря брака, вам нужно быть в браке."))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	s := shop.New(&shop.Opts{
		B:        b,
		Params:   params,
		Bot:      bot,
		Manager:  c.manager,
		Cm:       c.cm,
		Log:      c.log,
		BrakRepo: c.brakRepo,
	})

	_, err := bot.SendMessage(params.
		WithText(s.Label).
		WithReplyMarkup(&telego.InlineKeyboardMarkup{
			InlineKeyboard: s.ShopCallbacks,
		}),
	)
	if err != nil {
		c.log.Sugar().Error(err)
	}
}
