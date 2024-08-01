package family

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
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

	text := fmt.Sprintf("🍞🍞 %s 🍞🍞\n", html.Bold("Профиль"))
	text += fmt.Sprintf("👤 %s\n", fUser.Mention())
	messageCount, err := p.messages.MessageCount(from.ID, update.Message.Chat.ID)
	if err == nil {
		text += fmt.Sprintf("💬 %v\n", messageCount)
	}

	keyboard := tu.InlineKeyboardRow()

	b, _ := p.braks.FindByUserID(from.ID)

	if b != nil {
		if b.ChatID == 0 && update.Message.Chat.Type != "private" {
			b.ChatID = update.Message.Chat.ID
			_ = p.braks.Update(bson.M{"_id": b.OID}, bson.M{"$set": bson.M{"chat_id": b.ChatID}})
		}
		tUser, _ := p.users.FindByID(b.PartnerID(fUser.ID))
		if tUser == nil {
			return
		}
		keyboard = append(keyboard, tu.InlineKeyboardButton("🎰").WithCallbackData(static.CasinoData))
		keyboard = append(keyboard, tu.InlineKeyboardButton("🐹").WithCallbackData(static.HamsterData))

		text += fmt.Sprintf("\n❤️‍🔥❤️‍🔥      %s      ️‍❤️‍🔥❤️‍🔥\n", html.Bold("Брак"))
		text += fmt.Sprintf("🫂 %s [%s]\n", tUser.Mention(), b.Duration())

		if b.BabyUserID != nil {
			keyboard = append(keyboard, tu.InlineKeyboardButton("🍼").WithCallbackData(static.GrowKidData))
			bUser, err := p.users.FindByID(*b.BabyUserID)
			if err == nil {
				text += fmt.Sprintf("👼 %s [%s]\n", bUser.Mention(), b.DurationKid())
			}
		}

		text += fmt.Sprintf("💰 %v\n", b.Score)
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
