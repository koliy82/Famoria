package payments

import (
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/checkout"
	"famoria/internal/database/mongo/repositories/payment"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type Opts struct {
	fx.In
	Bh           *th.BotHandler
	Log          *zap.Logger
	UserRepo     user.Repository
	BrakRepo     brak.Repository
	CheckoutRepo checkout.Repository
	PaymentRepo  payment.Repository
}

func Register(opts Opts) {
	opts.Bh.HandlePreCheckoutQuery(func(bot *telego.Bot, update telego.PreCheckoutQuery) {
		params := &telego.AnswerPreCheckoutQueryParams{
			PreCheckoutQueryID: update.ID,
			Ok:                 false,
		}
		from, err := opts.UserRepo.FindOrUpdate(&update.From)
		if err != nil {
			err = bot.AnswerPreCheckoutQuery(params.WithErrorMessage("Брак не найден, покупка отменена."))
			if err != nil {
				opts.Log.Sugar().Error(err)
			}
			return
		}
		err = opts.CheckoutRepo.Insert(&checkout.Checkout{
			ID:               update.ID,
			FromId:           update.From.ID,
			From:             from,
			Currency:         update.Currency,
			TotalAmount:      update.TotalAmount,
			InvoicePayload:   update.InvoicePayload,
			ShippingOptionID: &update.ShippingOptionID,
		})
		if err != nil {
			opts.Log.Sugar().Error(err)
			err = bot.AnswerPreCheckoutQuery(params.WithErrorMessage("Произошла ошибка при обработке вашего платежа: " + err.Error()))
			return
		}
		err = bot.AnswerPreCheckoutQuery(params.WithOk())
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
		opts.Log.Sugar().Info("PreCheckoutQuery: ", zap.Any("update", update))
	})

	opts.Bh.Handle(func(bot *telego.Bot, update telego.Update) {
		m := update.Message
		err := opts.PaymentRepo.Insert(m)
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
		switch m.SuccessfulPayment.InvoicePayload {
		case Sub30:
			params := &telego.SendMessageParams{
				ChatID:    tu.ID(update.Message.Chat.ID),
				ParseMode: telego.ModeHTML,
			}
			b, err := opts.BrakRepo.FindByUserID(m.From.ID, nil)
			if err != nil {
				_, _ = bot.SendMessage(params.WithText("🚫 Ошибка при получении брака. Свяжитесь с администратором бота для решения проблемы."))
				opts.Log.Sugar().Error("#Subscribe Payment Error (get brak), user_id: " + strconv.FormatInt(m.From.ID, 10))
				opts.Log.Sugar().Error(err)
				return
			}
			b.AddSubDays(time.Duration(30) * time.Hour * 24)
			err = opts.BrakRepo.Update(
				bson.M{"_id": b.OID},
				bson.M{"$set": bson.M{"subscribe_end": b.SubscribeEnd}},
			)
			if err != nil {
				_, _ = bot.SendMessage(params.WithText("🚫 Ошибка при обновлении подписки. Свяжитесь с администратором бота для решения проблемы."))
				opts.Log.Sugar().Error("#Subscribe Payment Error (add 30 days), user_id: " + strconv.FormatInt(m.From.ID, 10))
				opts.Log.Sugar().Error(err)
				return
			}
			fUser, err := opts.UserRepo.FindByID(m.From.ID)
			if err != nil {
				_, _ = bot.SendMessage(params.WithText("Вы успешно приобрели подписку на 30 дней для брака."))
			}
			sUser, err := opts.UserRepo.FindByID(b.PartnerID(m.From.ID))
			if err != nil {
				_, _ = bot.SendMessage(params.WithText("Вы успешно приобрели подписку на 30 дней для брака."))
			}
			_, _ = bot.SendMessage(params.WithText(fmt.Sprintf(
				"%s, вы успешно приобрели подписку на 30 дней для брака с %s.",
				html.ModelMention(fUser), html.ModelMention(sUser),
			)))
		}
	}, th.SuccessPayment())
}
