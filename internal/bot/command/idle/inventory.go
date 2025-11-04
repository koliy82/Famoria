package idle

import (
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/inventory/showInventory"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type inventoryCmd struct {
	brakRepo brak.Repository
	userRepo user.Repository
	cm       *callback.CallbacksManager
	log      *zap.Logger
	manager  *item.Manager
}

func (c inventoryCmd) Handle(ctx *th.Context, update telego.Update) error {
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
		_, err := ctx.Bot().SendMessage(context.Background(), params.WithText("Для просмотра инвентаря брака, вам нужно быть в браке."))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return err
	}
	pages := showInventory.New(&showInventory.Opts{
		B:              b,
		Params:         params,
		BotCtx:         ctx,
		Manager:        c.manager,
		Log:            c.log,
		Cm:             c.cm,
		InventoryItems: b.Inventory.Items,
	})
	if pages == nil {
		return fmt.Errorf("pages is null")
	}

	_, err := ctx.Bot().SendMessage(context.Background(), params.
		WithText(pages.Label+"Выберите предмет для просмотра.\n").
		WithReplyMarkup(&telego.InlineKeyboardMarkup{
			InlineKeyboard: pages.ShowCallbacks,
		}),
	)

	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
