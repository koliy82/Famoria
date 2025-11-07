package family

import (
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/callback/static"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/message"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/common/buttons"
	"famoria/internal/pkg/html"
	"famoria/internal/pkg/plural"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
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

func (c profileCmd) Handle(ctx *th.Context, update telego.Update) error {
	from := update.Message.From
	fUser, err := c.userRepo.FindOrUpdate(from)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("🍞🍞 %s 🍞🍞\n", html.Bold("Профиль"))
	text += fmt.Sprintf("👤 %s\n", html.CodeInline(fUser.UsernameOrFull()))
	text += fmt.Sprintf("💰 %s\n", fUser.Score.GetFormattedScore())

	messageCount, err := c.messageRepo.MessageCount(from.ID, update.Message.Chat.ID)
	if err == nil {
		text += fmt.Sprintf("💬 %v\n", messageCount)
	}

	keyboard := buttons.New(5, 3)

	b, err := c.brakRepo.FindByUserID(from.ID, nil)
	//if err != nil {
	//	c.log.Sugar().Error("Брак не найден:", err)
	//}
	if b != nil {
		if b.ChatID == 0 && update.Message.Chat.Type != "private" {
			b.ChatID = update.Message.Chat.ID
			err = c.brakRepo.Update(bson.M{"_id": b.OID}, bson.M{"$set": bson.M{"chat_id": b.ChatID}})
			if err != nil {
				c.log.Sugar().Error(err)
				return err
			}
		}

		keyboard.Add(tu.InlineKeyboardButton("🎰").WithCallbackData(static.CasinoData))
		keyboard.Add(tu.InlineKeyboardButton("🐹").WithCallbackData(static.HamsterData))

		tUser, _ := c.userRepo.FindByID(b.PartnerID(fUser.ID))
		text += fmt.Sprintf("\n❤️‍🔥❤️‍🔥      %s      ️‍❤️‍🔥❤️‍🔥\n", html.Bold("Брак"))
		if tUser != nil {
			text += fmt.Sprintf("🫂 %s [%s]\n", html.CodeInline(tUser.UsernameOrFull()), b.Duration())
		}

		if b.BabyUserID != nil {
			keyboard.Add(tu.InlineKeyboardButton("🍼").WithCallbackData(static.GrowKidData))
			bUser, err := c.userRepo.FindByID(*b.BabyUserID)
			if err == nil {
				text += fmt.Sprintf("👼 %s [%s]\n", html.CodeInline(bUser.UsernameOrFull()), b.DurationKid())
			}
		}

		if b.IsSub() {
			days := b.SubDaysCount()
			text += html.Bold(fmt.Sprintf("💎 Подписка на %s\n", fmt.Sprintf("%v %s", days, plural.Declension(days, "день", "дня", "дней"))))
			keyboard.Add(tu.InlineKeyboardButton("💎").WithCallbackData(static.AnubisData))
		} else {
			text += fmt.Sprintf("😿 Нет активной подписки\n")
		}

		text += fmt.Sprintf("💰 %v\n", b.Score.GetFormattedScore())
	}

	params := &telego.SendMessageParams{
		ChatID:              tu.ID(update.Message.Chat.ID),
		ParseMode:           telego.ModeHTML,
		Text:                text,
		DisableNotification: true,
	}

	if len(keyboard.Buttons) != 0 {
		params.ReplyMarkup = keyboard.Build()
	}

	_, err = ctx.Bot().SendMessage(context.Background(), params)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
