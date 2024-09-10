package donate

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/handler/payments"
	"famoria/internal/bot/idle/item"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"time"
)

type SubscribeCmd struct {
	brakRepo brak.Repository
	userRepo user.Repository
	log      *zap.Logger
	cm       *callback.CallbacksManager
	m        *item.Manager
}

func (c SubscribeCmd) Handle(bot *telego.Bot, update telego.Update) {
	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.GetMessageID(),
			AllowSendingWithoutReply: true,
		},
	}
	b, err := c.brakRepo.FindByUserID(update.Message.From.ID, c.m)
	if err != nil {
		_, err := bot.SendMessage(params.WithText("🚫 Вы не состоите в браке, подписка покупается на действующий брак. Женитесь пожалуйста командой /gobrak. 🥺"))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}
	fUser, err := c.userRepo.FindByID(b.FirstUserID)
	if err != nil {
		fUser = &user.User{
			ID:        update.Message.From.ID,
			FirstName: "?",
		}
	}
	sUser, err := c.userRepo.FindByID(b.SecondUserID)
	if err != nil {
		sUser = &user.User{
			ID:        b.SecondUserID,
			FirstName: "?",
		}
	}
	fUser.UsernameOrFull()
	s30Callback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "⭐️ Telegram Stars",
		CtxType:  callback.OneClick,
		OwnerIDs: []int64{b.FirstUserID, b.SecondUserID},
		Time:     time.Duration(1) * time.Hour,
		Callback: func(query telego.CallbackQuery) {
			invoice, err := bot.SendInvoice(&telego.SendInvoiceParams{
				ChatID: params.ChatID,
				Title:  "Famoria - подписка на 30 дней.",
				Description: fmt.Sprintf(
					"Подписка для брака %s и %s.",
					fUser.UsernameOrFull(), sUser.UsernameOrFull(),
				),
				Payload:  payments.Sub30,
				Currency: "XTR",
				Prices: []telego.LabeledPrice{
					{
						Label:  "30 дней",
						Amount: 82,
					},
				},
				//StartParameter:            "",
				PhotoURL: "https://i.ytimg.com/vi/NVcPeHtxLNE/maxresdefault.jpg",
				//PhotoSize:                 0,
				//PhotoWidth:                0,
				//PhotoHeight:               0,
				DisableNotification: false,
				ProtectContent:      false,
				//MessageEffectID:           "",
				ReplyParameters: params.ReplyParameters,
				//ReplyMarkup:               nil,
			})
			if err != nil {
				c.log.Sugar().Error(err)
				return
			}
			c.log.Sugar().Info(invoice)
		},
	})
	text := "Famoria - подписка, дающая следующие преимужества:\n"

	body := "+ 20% больше монет с любых источников дохода.\n"
	body += "+ В топе отображается с эмодзи.\n"
	body += "+ 20% скидка в потайной лавке.\n"
	body += "+ Доступ к премиум-игре Анубис.\n"
	body += "+ Действует на обоих участников брака.\n"
	text += html.CodeBlockWithLang(body, "Subscription buffs")
	text += html.Italic("Помогает оплачивать хостинг боту.")
	_, err = bot.SendMessage(params.WithText(text).
		WithReplyMarkup(tu.InlineKeyboard(
			tu.InlineKeyboardRow(s30Callback.Inline()),
		)),
	)
	if err != nil {
		c.log.Sugar().Error(err)
	}
}
