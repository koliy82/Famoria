package family

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
	"go_tg_bot/internal/utils/html"
	"time"
)

type endFamily struct {
	cm    *callback.CallbacksManager
	braks brak.Repository
	users user.Repository
}

func (e endFamily) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	brak, _ := e.braks.FindByUserID(from.ID)

	if brak == nil {
		_, _ = bot.SendMessage(&telego.SendMessageParams{
			ChatID:    tu.ID(update.Message.Chat.ID),
			ParseMode: telego.ModeHTML,
			Text:      fmt.Sprintf("%s, ты не состоишь в браке. 😥", html.UserMention(from)),
		})
		return
	}

	yesCallback := e.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Да.",
		CtxType:  callback.OneClick,
		OwnerIDs: []int64{from.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			err := e.braks.Delete(brak.BID)
			if err != nil {
				_, _ = bot.SendMessage(&telego.SendMessageParams{
					ChatID:    tu.ID(update.Message.Chat.ID),
					ParseMode: telego.ModeHTML,
					Text:      fmt.Sprintf("%s, произошла ошибка при разводе. 😥", html.UserMention(from)),
				})
				return
			}
			fuser, err := e.users.FindByID(brak.FirstUserID)
			if err != nil {
				return
			}
			tuser, err := e.users.FindByID(brak.SecondUserID)
			if err != nil {
				return
			}
			_, _ = bot.SendMessage(&telego.SendMessageParams{
				ChatID:    tu.ID(update.Message.Chat.ID),
				ParseMode: telego.ModeHTML,
				Text: fmt.Sprintf(
					"Брак между %s и %s распался. 💔\nОни прожили вместе %s",
					fuser.Mention(), tuser.Mention(), brak.CreateDate.String(),
				),
				ReplyParameters: &telego.ReplyParameters{
					MessageID: update.Message.GetMessageID(),
				},
			})
		},
	})

	_, _ = bot.SendMessage(&telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
		Text:      fmt.Sprintf("%s, ты уверен? 💔", html.UserMention(from)),
		ReplyMarkup: tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				yesCallback.Inline(),
			),
		),
	})
}
