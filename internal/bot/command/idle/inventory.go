package idle

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	items2 "famoria/internal/bot/idle/item/items"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"time"
)

type inventoryCmd struct {
	brakRepo brak.Repository
	userRepo user.Repository
	cm       *callback.CallbacksManager
	log      *zap.Logger
	manager  *item.Manager
}

func (c inventoryCmd) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	b, _ := c.brakRepo.FindByUserID(from.ID, nil)
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
	items := b.Inventory.GetItems(c.manager)
	if len(items) == 0 {
		_, err := bot.SendMessage(params.WithText("Инвентарь брака пуст, милорд."))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}
	//var rows [][]telego.InlineKeyboardButton
	var callbacks [][]telego.InlineKeyboardButton
	chunkSize := 5
	var lastSelectedItem *items2.Name
	for i := 0; i < len(items); i += chunkSize {
		end := i + chunkSize
		if end > len(items) {
			end = len(items)
		}
		current := items[i:end]
		//rows = append(rows, make([]telego.InlineKeyboardButton, len(current)))
		//for j, item := range current {
		//	rows[len(rows)-1][j] = item.ToButton()
		//}
		callbacks = append(callbacks, make([]telego.InlineKeyboardButton, len(current)))
		for j, si := range current {
			dCallback := c.cm.DynamicCallback(callback.DynamicOpts{
				Label:    si.Name.String(),
				CtxType:  callback.Temporary,
				OwnerIDs: []int64{b.FirstUserID, b.SecondUserID},
				Time:     time.Duration(30) * time.Minute,
				Callback: func(query telego.CallbackQuery) {
					if lastSelectedItem != nil && *lastSelectedItem == si.Name {
						return
					}
					lastSelectedItem = &si.Name
					_, err := bot.EditMessageText(&telego.EditMessageTextParams{
						MessageID: query.Message.GetMessageID(),
						ChatID:    tu.ID(update.Message.Chat.ID),
						ParseMode: telego.ModeHTML,
						Text:      si.FullDescription(),
						ReplyMarkup: &telego.InlineKeyboardMarkup{
							InlineKeyboard: callbacks,
						},
					})
					if err != nil {
						c.log.Sugar().Error(err)
					}
				},
			})
			callbacks[len(callbacks)-1][j] = dCallback.Inline()
		}
	}

	_, err := bot.SendMessage(params.
		WithText("Выберите предмет для просмотра.").
		WithReplyMarkup(&telego.InlineKeyboardMarkup{
			InlineKeyboard: callbacks,
		}),
	)

	if err != nil {
		c.log.Sugar().Error(err)
	}
}
