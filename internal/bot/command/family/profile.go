package family

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/callback/static"
	"famoria/internal/database/clickhouse/repositories/message"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/html"
	"famoria/internal/pkg/plural"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type profileCmd struct {
	cm          *callback.CallbacksManager
	log         *zap.Logger
	userRepo    user.Repository
	brakRepo    brak.Repository
	messageRepo message.Repository
}

func (c profileCmd) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	fUser, err := c.userRepo.FindOrUpdate(from)
	if err != nil {
		return
	}

	text := fmt.Sprintf("🍞🍞 %s 🍞🍞\n", html.Bold("Профиль"))
	text += fmt.Sprintf("👤 %s\n", html.CodeInline(fUser.UsernameOrFull()))
	text += fmt.Sprintf("💰 %s\n", fUser.Score.GetFormattedScore())

	//text += fmt.Sprintf("old💰 %s\n", fUser.Score.GetFormattedScore())
	//for range 1 {
	//	fUser.Score.Increase(1000)
	//}
	//_ = c.userRepo.Update(bson.M{"id": fUser.ID}, bson.M{"$set": bson.M{"score": fUser.Score}})
	//text += fmt.Sprintf("new💰 %s\n", fUser.Score.GetFormattedScore())

	messageCount, err := c.messageRepo.MessageCount(from.ID, update.Message.Chat.ID)
	if err == nil {
		text += fmt.Sprintf("💬 %v\n", messageCount)
	}

	if fUser.IsSub() {
		days := fUser.SubDaysCount()
		text += html.Bold(fmt.Sprintf("💎 %s\n", fmt.Sprintf("%v %s", days, plural.Declension(days, "день", "дня", "дней"))))
	} else {
		text += fmt.Sprintf("😿 Нет активной подписки\n")
	}

	keyboard := tu.InlineKeyboardRow()

	b, _ := c.brakRepo.FindByUserID(from.ID)

	if b != nil {
		if b.ChatID == 0 && update.Message.Chat.Type != "private" {
			b.ChatID = update.Message.Chat.ID
			err = c.brakRepo.Update(bson.M{"_id": b.OID}, bson.M{"$set": bson.M{"chat_id": b.ChatID}})
			if err != nil {
				c.log.Sugar().Error(err)
				return
			}
		}

		keyboard = append(keyboard, tu.InlineKeyboardButton("🎰").WithCallbackData(static.CasinoData))
		keyboard = append(keyboard, tu.InlineKeyboardButton("🐹").WithCallbackData(static.HamsterData))

		tUser, _ := c.userRepo.FindByID(b.PartnerID(fUser.ID))
		text += fmt.Sprintf("\n❤️‍🔥❤️‍🔥      %s      ️‍❤️‍🔥❤️‍🔥\n", html.Bold("Брак"))
		if tUser != nil {
			text += fmt.Sprintf("🫂 %s [%s]\n", html.CodeInline(tUser.UsernameOrFull()), b.Duration())
		}

		if b.BabyUserID != nil {
			keyboard = append(keyboard, tu.InlineKeyboardButton("🍼").WithCallbackData(static.GrowKidData))
			bUser, err := c.userRepo.FindByID(*b.BabyUserID)
			if err == nil {
				text += fmt.Sprintf("👼 %s [%s]\n", html.CodeInline(bUser.UsernameOrFull()), b.DurationKid())
			}
		}

		text += fmt.Sprintf("💰 %v\n", b.Score.GetFormattedScore())

		text += fmt.Sprintf("items: %v\n", len(b.Inventory.Items))
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

	_, err = bot.SendMessage(params)
	if err != nil {
		c.log.Sugar().Error(err)
	}
}
