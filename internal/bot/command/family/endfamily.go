package family

import (
	"famoria/internal/bot/callback"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/koliy82/telego"
	tu "github.com/koliy82/telego/telegoutil"
	"go.uber.org/zap"
	"time"
)

type endFamily struct {
	cm       *callback.CallbacksManager
	log      *zap.Logger
	brakRepo brak.Repository
	userRepo user.Repository
}

func (e endFamily) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	brak, _ := e.brakRepo.FindByUserID(from.ID)
	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
	}

	if brak == nil {
		_, err := bot.SendMessage(params.
			WithText(fmt.Sprintf("%s, ты не состоишь в браке. 😥", html.UserMention(from))),
		)
		if err != nil {
			e.log.Sugar().Error(err)
		}
		return
	}

	yesCallback := e.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Да.",
		CtxType:  callback.OneClick,
		OwnerIDs: []int64{from.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			err := e.brakRepo.Delete(brak.OID)
			if err != nil {
				_, err := bot.SendMessage(params.
					WithText(fmt.Sprintf("%s, произошла ошибка при разводе. 😥", html.UserMention(from))).
					WithReplyMarkup(nil),
				)
				if err != nil {
					e.log.Sugar().Error(err)
				}
				return
			}
			fuser, err := e.userRepo.FindByID(brak.FirstUserID)
			if err != nil {
				return
			}
			tuser, err := e.userRepo.FindByID(brak.SecondUserID)
			if err != nil {
				return
			}
			_, err = bot.SendMessage(params.
				WithText(fmt.Sprintf(
					"Брак между %s и %s распался. 💔\nОни прожили вместе %s",
					html.ModelMention(fuser), html.ModelMention(tuser), brak.Duration(),
				)).WithReplyMarkup(nil),
			)
			if err != nil {
				e.log.Sugar().Error(err)
			}
		},
	})

	_, err := bot.SendMessage(params.
		WithText(fmt.Sprintf("%s, ты уверен? 💔", html.UserMention(from))).
		WithReplyMarkup(tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				yesCallback.Inline(),
			),
		)),
	)
	if err != nil {
		e.log.Sugar().Error(err)
	}
}
