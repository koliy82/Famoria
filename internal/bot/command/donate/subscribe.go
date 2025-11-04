package donate

import (
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/handler/payments"
	"famoria/internal/bot/idle/item"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/common/buttons"
	"famoria/internal/pkg/html"
	"fmt"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type SubscribeCmd struct {
	brakRepo    brak.Repository
	userRepo    user.Repository
	log         *zap.Logger
	cm          *callback.CallbacksManager
	m           *item.Manager
	yKassaToken *string
}

func (c SubscribeCmd) Handle(ctx *th.Context, update telego.Update) error {
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
		_, err := ctx.Bot().SendMessage(context.Background(), params.WithText("🚫 Вы не состоите в браке, подписка покупается на действующий брак. Женитесь пожалуйста командой /gobrak. 🥺"))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return err
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

	builder := buttons.New(5, 1)
	starsCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "⭐️ Telegram Stars",
		CtxType:  callback.OneClick,
		OwnerIDs: []int64{b.FirstUserID, b.SecondUserID},
		Time:     time.Duration(1) * time.Hour,
		Callback: func(query telego.CallbackQuery) {
			invoice, err := ctx.Bot().SendInvoice(context.Background(), &telego.SendInvoiceParams{
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
	builder.Add(starsCallback.Inline())

	if c.yKassaToken != nil {
		data := common.ProviderData{
			Receipt: common.Receipt{
				Items: []common.Item{
					{
						Description: "Игровая подписка на Telegram-бота (30 дней)",
						Quantity:    1,
						Amount: common.Amount{
							Currency: "RUB",
							Value:    "139.00",
						},
						VatCode: 1,
					},
				},
			},
		}
		yooKCallback := c.cm.DynamicCallback(callback.DynamicOpts{
			Label:    "🇷🇺 ЮKassa",
			CtxType:  callback.OneClick,
			OwnerIDs: []int64{b.FirstUserID, b.SecondUserID},
			Time:     time.Duration(1) * time.Hour,
			Callback: func(query telego.CallbackQuery) {
				invoice, err := ctx.Bot().SendInvoice(context.Background(), &telego.SendInvoiceParams{
					ChatID: params.ChatID,
					Title:  "Famoria - подписка на 30 дней.",
					Description: fmt.Sprintf(
						"Подписка для брака %s и %s.",
						fUser.UsernameOrFull(), sUser.UsernameOrFull(),
					),
					Payload:  payments.Sub30,
					Currency: "RUB",
					Prices: []telego.LabeledPrice{
						{
							Label:  "30 дней",
							Amount: 13900,
						},
					},
					NeedEmail:           true,
					SendEmailToProvider: true,
					ProviderToken:       *c.yKassaToken,
					ProviderData:        data.ToJson(),
					PhotoURL:            "https://i.ytimg.com/vi/QFYpp-cpy9w/hq720.jpg?sqp=-oaymwEhCK4FEIIDSFryq4qpAxMIARUAAAAAGAElAADIQj0AgKJD&rs=AOn4CLCWdu-QiXAtWE67vOH-7FEldF6KFw",
					DisableNotification: false,
					ProtectContent:      false,
					ReplyParameters:     params.ReplyParameters,
				})
				if err != nil {
					c.log.Sugar().Error(err)
					return
				}
				c.log.Sugar().Info(invoice)
			},
		})
		builder.Add(yooKCallback.Inline())
	}

	text := "Famoria - подписка за 139₽ <s>(459₽)</s>, дающая следующие преимужества:\n"
	body := "+ 20% больше монет с любых источников дохода.\n"
	body += "+ 20% скидка в потайной лавке.\n"
	body += "+ Действует на обоих участников брака.\n"
	body += "+ В топе отображается с эмодзи.\n"
	body += "+ Доступ к премиум-игре Анубис:\n"
	body += "  - 3 попытки в день.\n"
	body += "  - 1000 базовой силы.\n"
	body += "  - 75% шанс на победу.\n"
	body += "  - 1% на x20 выйгрыша.\n"
	body += "  - 1% умножения счёта на 20%.\n"
	text += html.CodeBlockWithLang(body, "Subscription buffs")
	text += html.Italic("Помогает оплачивать хостинг боту.")
	_, err = ctx.Bot().SendMessage(context.Background(),
		params.WithText(text).WithReplyMarkup(builder.Build()),
	)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
