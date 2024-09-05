package family

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/events"
	"famoria/internal/bot/idle/events/casino"
	"famoria/internal/bot/idle/events/growkid"
	"famoria/internal/bot/idle/events/hamster"
	"famoria/internal/bot/idle/item/inventory"
	"famoria/internal/bot/idle/item/items"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"time"
)

type goFamilyCmd struct {
	cm       *callback.CallbacksManager
	brakRepo brak.Repository
	log      *zap.Logger
}

func (c goFamilyCmd) Handle(bot *telego.Bot, update telego.Update) {
	fUser := update.Message.From
	reply := update.Message.ReplyToMessage

	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.GetMessageID(),
			AllowSendingWithoutReply: true,
		},
	}

	if reply == nil {
		_, err := bot.SendMessage(params.
			WithText(fmt.Sprintf(
				"%s, ответь на любое сообщение партнёра. 😘💬",
				html.UserMention(fUser),
			)))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	tUser := reply.From
	if tUser.ID == fUser.ID {
		_, err := bot.SendMessage(params.WithText(fmt.Sprintf(
			"%s, брак с собой нельзя, придётся искать пару. 😥",
			html.UserMention(fUser),
		)))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	if tUser.IsBot {
		_, err := bot.SendMessage(params.WithText(fmt.Sprintf(
			"%s, бота не трогай. 👿",
			html.UserMention(fUser),
		)))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	fBrakCount, _ := c.brakRepo.Count(bson.M{"$or": []interface{}{
		bson.M{"first_user_id": fUser.ID},
		bson.M{"second_user_id": fUser.ID},
	}})
	if fBrakCount != 0 {
		_, err := bot.SendMessage(params.WithText(fmt.Sprintf(
			"%s, у вас уже есть брак! 💍",
			html.UserMention(fUser),
		)))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	tBrakCount, _ := c.brakRepo.Count(bson.M{"$or": []interface{}{
		bson.M{"first_user_id": tUser.ID},
		bson.M{"second_user_id": tUser.ID},
	}})
	if tBrakCount != 0 {
		_, err := bot.SendMessage(params.WithText(fmt.Sprintf(
			"%s, у вашего партнёра уже есть брак! 💍",
			html.UserMention(fUser),
		)))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	yesCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Да!❤️‍🔥",
		CtxType:  callback.ChooseOne,
		OwnerIDs: []int64{tUser.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_ = c.brakRepo.Insert(&brak.Brak{
				OID:          primitive.NewObjectID(),
				ChatID:       update.Message.Chat.ID,
				FirstUserID:  fUser.ID,
				SecondUserID: tUser.ID,
				CreateDate:   time.Now(),
				Inventory:    &inventory.Inventory{Items: make(map[items.Name]inventory.Item)},
				Score: common.Score{
					Mantissa: 0,
					Exponent: 0,
				},
				Hamster: &hamster.Hamster{
					Base: events.Base{
						LastPlay:  time.Time{},
						PlayCount: 0,
					},
				},
				Casino: &casino.Casino{
					Base: events.Base{
						LastPlay:  time.Time{},
						PlayCount: 0,
					},
				},
				GrowKid: &growkid.GrowKid{
					Base: events.Base{
						LastPlay:  time.Time{},
						PlayCount: 0,
					},
				},
			})

			_, err := bot.SendMessage(&telego.SendMessageParams{
				ChatID:    tu.ID(update.Message.Chat.ID),
				ParseMode: telego.ModeHTML,
				Text: fmt.Sprintf(
					"Внимание! ⚠️\n%s и %s теперь вместе ❤️‍🔥",
					html.UserMention(fUser), html.UserMention(tUser),
				),
				ReplyParameters: &telego.ReplyParameters{
					MessageID: query.Message.GetMessageID(),
				},
			})
			if err != nil {
				c.log.Sugar().Error(err)
			}
		},
	})

	noCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:      "Нет!💔",
		CtxType:    callback.ChooseOne,
		OwnerIDs:   []int64{tUser.ID},
		Time:       time.Duration(60) * time.Minute,
		AnswerText: "Отказ 🖤",
		Callback: func(query telego.CallbackQuery) {
			_, err := bot.SendMessage(&telego.SendMessageParams{
				ChatID: tu.ID(update.Message.Chat.ID),
				Text:   "Отказ 🖤",
				ReplyParameters: &telego.ReplyParameters{
					MessageID: query.Message.GetMessageID(),
				},
			})
			if err != nil {
				c.log.Sugar().Error(err)
				return
			}
		},
	})

	_, err := bot.SendMessage(params.WithText(fmt.Sprintf(
		"💍 %s, минуточку внимания.\n"+
			"💖 %s сделал вам предложение руки и сердца.",
		html.UserMention(tUser), html.UserMention(fUser),
	)).WithReplyMarkup(tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			yesCallback.Inline(),
			noCallback.Inline(),
		),
	)))
	if err != nil {
		c.log.Sugar().Error(err)
	}

}
