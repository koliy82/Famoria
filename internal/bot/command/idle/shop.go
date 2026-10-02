package idle

import (
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/shop"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/i18n"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type shopCmd struct {
	brakRepo     brak.Repository
	userRepo     user.Repository
	chatSettings chat_settings.Repository
	cm           *callback.CallbacksManager
	log          *zap.Logger
	manager      *item.Manager
}

func (c shopCmd) Handle(ctx *th.Context, update telego.Update) error {
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
		lang := c.chatSettings.Lang(update.Message.Chat.ID)
		_, err := ctx.Bot().SendMessage(context.Background(), params.WithText(i18n.T(lang, i18n.KeyShopNeedMarriage)))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return err
	}

	s, err := shop.New(&shop.Opts{
		B:        b,
		Params:   params,
		BotCtx:   ctx,
		Manager:  c.manager,
		Cm:       c.cm,
		Log:      c.log,
		BrakRepo: c.brakRepo,
	})
	if err == nil {
		_, err = ctx.Bot().SendMessage(context.Background(), params.
			WithText(s.Label).
			WithReplyMarkup(&telego.InlineKeyboardMarkup{
				InlineKeyboard: s.ShopCallbacks,
			}),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
	}
	return err
}
