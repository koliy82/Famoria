package family

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/bot/callback/static"
	"go_tg_bot/internal/database/clickhouse/repositories/message"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
	"go_tg_bot/internal/utils/html"
)

type profile struct {
	cm       *callback.CallbacksManager
	log      *zap.Logger
	users    user.Repository
	braks    brak.Repository
	messages message.Repository
}

func (p profile) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	fUser, err := p.users.FindByID(from.ID)
	if err != nil {
		return
	}

	text := fmt.Sprintf("🍞🍞🍞 %s 🍞🍞🍞\n", html.Bold("Профиль"))
	text += fmt.Sprintf("%s\n", fUser.Mention())
	text += fmt.Sprintf("Хинкали: %v\n", fUser.MessageCount)
	messageCount, err := p.messages.MessageCount(from.ID, update.Message.Chat.ID)
	if err == nil {
		text += fmt.Sprintf("Сообщений в чате: %v\n", messageCount)
	}
	text += fmt.Sprintf("Сообщений всего: %v\n", fUser.MessageCount)

	keyboard := tu.InlineKeyboardRow()

	b, _ := p.braks.FindByUserID(from.ID)
	tUser, err := p.users.FindByID(b.PartnerID(fUser.ID))
	if b != nil && tUser != nil {
		score := fUser.MessageCount + uint64(b.Score)
		keyboard = append(keyboard, tu.InlineKeyboardButton("🧊").WithCallbackData(static.CasinoData))

		text += fmt.Sprintf("\n❤️‍🔥👨🏻‍🦱❤️‍🔥 %s ❤️‍🔥👩🏻‍🦱❤️‍🔥\n", html.Bold("Партнёр"))
		text += fmt.Sprintf("%s\n", tUser.Mention())

		if b.BabyUserID != nil {
			keyboard = append(keyboard, tu.InlineKeyboardButton("👶🏻").WithCallbackData(static.GrowKidData))
			bUser, err := p.users.FindByID(*b.BabyUserID)
			if err == nil {
				text += fmt.Sprintf("Ребёнок: %s\n", bUser.Mention())
			}
		}

		text += fmt.Sprintf("Вместе: %s\n", b.Duration())
		text += fmt.Sprintf("Хинкали: %v\n", score)
	}

	params := &telego.SendMessageParams{
		ChatID:              tu.ID(update.Message.Chat.ID),
		ParseMode:           telego.ModeHTML,
		Text:                text,
		DisableNotification: true,
	}

	if len(keyboard) != 0 {
		params.ReplyMarkup = tu.InlineKeyboard(keyboard)
	}

	_, _ = bot.SendMessage(params)
}
