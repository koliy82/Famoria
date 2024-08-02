package family

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
	"go_tg_bot/internal/pkg/html"
	"time"
)

type endKid struct {
	cm       *callback.CallbacksManager
	brakRepo brak.Repository
	userRepo user.Repository
	log      *zap.Logger
}

func (e endKid) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	b, _ := e.brakRepo.FindByUserID(from.ID)

	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
	}

	if b == nil {
		_, err := bot.SendMessage(params.
			WithText(fmt.Sprintf("%s, ты не состоишь в браке. 😥", html.UserMention(from))),
		)
		if err != nil {
			e.log.Sugar().Error(err)
		}
		return
	}

	if b.BabyUserID == nil {
		_, err := bot.SendMessage(params.
			WithText(fmt.Sprintf("%s, у вас нет детей. 🤔", html.UserMention(from))),
		)
		if err != nil {
			e.log.Sugar().Error(err)
		}
		return
	}

	sUser, _ := e.userRepo.FindByID(b.PartnerID(from.ID))
	if sUser == nil {
		return
	}
	bUser, _ := e.userRepo.FindByID(*b.BabyUserID)
	if bUser == nil {
		return
	}

	yesCallback := e.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Да.",
		CtxType:  callback.OneClick,
		OwnerIDs: []int64{sUser.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			err := e.brakRepo.Update(
				bson.M{"_id": b.OID},
				bson.M{"$set": bson.D{
					{"baby_user_id", nil},
					{"baby_create_date", nil},
				}},
			)
			if err != nil {
				e.log.Sugar().Error(err)
				return
			}

			_, err = bot.SendMessage(params.
				WithText(fmt.Sprintf("Внимание! ⚠️\n%s был аннигилирован %s и %s!\n Он прожил %s",
					html.ModelMention(bUser), html.UserMention(from), html.ModelMention(sUser), b.DurationKid())).
				WithReplyMarkup(nil),
			)
			if err != nil {
				e.log.Sugar().Error(err)
			}
		},
	})

	_, err := bot.SendMessage(params.
		WithText(fmt.Sprintf("%s, ты тоже хочешь аннигилировать %s? 😐",
			html.ModelMention(sUser), html.ModelMention(bUser))).
		WithReplyMarkup(tu.InlineKeyboard(
			tu.InlineKeyboardRow(yesCallback.Inline()),
		)),
	)
	if err != nil {
		e.log.Sugar().Error(err)
	}

}
