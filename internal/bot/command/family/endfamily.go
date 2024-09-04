package family

import (
	"famoria/internal/bot/callback"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"time"
)

type endFamilyCmd struct {
	cm       *callback.CallbacksManager
	log      *zap.Logger
	brakRepo brak.Repository
	userRepo user.Repository
}

func (c endFamilyCmd) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	b, _ := c.brakRepo.FindByUserID(from.ID, nil)
	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
	}

	if b == nil {
		_, err := bot.SendMessage(params.
			WithText(fmt.Sprintf("%s, ты не состоишь в браке. 😥", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	yesCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Да.",
		CtxType:  callback.OneClick,
		OwnerIDs: []int64{from.ID},
		Time:     time.Duration(60) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			err := c.brakRepo.Delete(b.OID)
			if err != nil {
				_, err := bot.SendMessage(params.
					WithText(fmt.Sprintf("%s, произошла ошибка при разводе. 😥", html.UserMention(from))).
					WithReplyMarkup(nil),
				)
				if err != nil {
					c.log.Sugar().Error(err)
				}
				return
			}
			fuser, err := c.userRepo.FindByID(b.FirstUserID)
			if err != nil {
				return
			}
			tuser, err := c.userRepo.FindByID(b.SecondUserID)
			if err != nil {
				return
			}
			_, err = bot.SendMessage(params.
				WithText(fmt.Sprintf(
					"Брак между %s и %s распался. 💔\nОни прожили вместе %s",
					html.ModelMention(fuser), html.ModelMention(tuser), b.Duration(),
				)).WithReplyMarkup(nil),
			)
			if err != nil {
				c.log.Sugar().Error(err)
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
		c.log.Sugar().Error(err)
	}
}
