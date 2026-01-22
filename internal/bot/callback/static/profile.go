package static

import (
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/event/anubis"
	"famoria/internal/bot/idle/event/casino"
	"famoria/internal/bot/idle/event/growkid"
	"famoria/internal/bot/idle/event/hamster"
	"famoria/internal/bot/idle/event/mining"
	"famoria/internal/bot/idle/item"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/html"
	"os"
	"strconv"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	GrowKidData = "grow_kid"
	CasinoData  = "casino"
	HamsterData = "hamster"
	AnubisData  = "anubis"
	MiningData  = "mining"
)

type Opts struct {
	fx.In
	Log      *zap.Logger
	BrakRepo brak.Repository
	UserRepo user.Repository
	Cm       *callback.CallbacksManager
	Bot      *telego.Bot
	M        *item.Manager
}

func ProfileCallbacks(opts Opts) {
	opts.Cm.StaticCallback(CasinoData, func(query telego.CallbackQuery) {
		b, err := opts.BrakRepo.FindByUserID(query.From.ID, opts.M)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для использования казино необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}

		response := b.Events.Casino.Play(&casino.PlayOpts{
			Log:   opts.Log,
			Bot:   opts.Bot,
			Query: query,
		})
		if response == nil {
			return
		}

		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$inc": bson.M{
					"score": response.Score,
				},
				"$set": bson.M{
					"events.casino": b.Events.Casino,
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error("Ошибка при обновлении счёта #casino (", response.Score, ") пользователя ", query.From.ID, ":", err)
			return
		}

		params := &telego.SendMessageParams{
			ChatID:    tu.ID(query.Message.GetChat().ID),
			ParseMode: telego.ModeHTML,
			Text:      response.Text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
		}

		if response.Path == "" {
			_, err = opts.Bot.SendMessage(context.Background(), params)
			if err != nil {
				opts.Log.Sugar().Error(err)
			}
		} else {
			gif, err := os.Open(response.Path)
			if err == nil {
				_, err = opts.Bot.SendAnimation(context.Background(), &telego.SendAnimationParams{
					Caption:   response.Text,
					ParseMode: telego.ModeHTML,
					ChatID:    params.ChatID,
					Animation: tu.File(gif),
				})
				err := gif.Close()
				if err != nil {
					opts.Log.Sugar().Error(err)
				}
			} else {
				opts.Log.Sugar().Error(err)
				_, err = opts.Bot.SendMessage(context.Background(), params)
			}
		}

		if err != nil {
			opts.Log.Sugar().Error(err)
		}
		err = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
	})

	opts.Cm.StaticCallback(GrowKidData, func(query telego.CallbackQuery) {
		b, err := opts.BrakRepo.FindByUserID(query.From.ID, opts.M)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для кормления ребёнка необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}
		if b.BabyUserID == nil {
			_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для кормления ребёнка его необходимо родить.",
				ShowAlert:       true,
			})
			return
		}

		response := b.Events.GrowKid.Play(&growkid.PlayOpts{
			Log:   opts.Log,
			Bot:   opts.Bot,
			Query: query,
		})
		if response == nil {
			return
		}

		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$inc": bson.M{
					"score": response.Score,
				},
				"$set": bson.M{
					"events.grow_kid": b.Events.GrowKid,
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error("Ошибка при обновлении счёта #grow_kid (", response.Score, ") пользователя ", query.From.ID, ":", err)
			return
		}
		_, _ = opts.Bot.SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID:    tu.ID(query.Message.GetChat().ID),
			ParseMode: telego.ModeHTML,
			Text:      response.Text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
		})
		_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
	})

	opts.Cm.StaticCallback(HamsterData, func(query telego.CallbackQuery) {
		b, err := opts.BrakRepo.FindByUserID(query.From.ID, opts.M)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для тапа хомяка необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}

		response := b.Events.Hamster.Play(&hamster.PlayOpts{
			Log:   opts.Log,
			Bot:   opts.Bot,
			Query: query,
		})
		if response == nil {
			return
		}
		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$inc": bson.M{
					"score": response.Score,
				},
				"$set": bson.M{
					"events.hamster": b.Events.Hamster,
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error("Ошибка при обновлении счёта #hamster (", response.Score, ") пользователя ", query.From.ID, ":", err)
			return
		}

		err = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "Успешный тап по хомяку +" + strconv.FormatInt(response.Score, 10),
		})
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
	})

	opts.Cm.StaticCallback(AnubisData, func(query telego.CallbackQuery) {
		b, err := opts.BrakRepo.FindByUserID(query.From.ID, opts.M)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для игры в анубис вы должны быть в браке и иметь действующую подписку.",
				ShowAlert:       true,
			})
			return
		}
		response := b.Events.Anubis.Play(&anubis.PlayOpts{
			Log:   opts.Log,
			Bot:   opts.Bot,
			Query: query,
			IsSub: b.IsSub(),
		})
		if response == nil {
			return
		}

		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$inc": bson.M{
					"score": response.Score,
				},
				"$set": bson.M{
					"events.anubis": b.Events.Anubis,
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error("Ошибка при обновлении счёта #anubis (", response.Score, ") пользователя ", query.From.ID, ":", err)
			return
		}

		params := &telego.SendMessageParams{
			ChatID:    tu.ID(query.Message.GetChat().ID),
			ParseMode: telego.ModeHTML,
			Text:      response.Text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
		}

		if response.Path == "" {
			_, err = opts.Bot.SendMessage(context.Background(), params)
			if err != nil {
				opts.Log.Sugar().Error(err)
			}
		} else {
			gif, err := os.Open(response.Path)
			if err == nil {
				_, err = opts.Bot.SendAnimation(context.Background(), &telego.SendAnimationParams{
					Caption:   response.Text,
					ParseMode: telego.ModeHTML,
					ChatID:    params.ChatID,
					Animation: tu.File(gif),
				})
				err := gif.Close()
				if err != nil {
					opts.Log.Sugar().Error(err)
				}
			} else {
				opts.Log.Sugar().Error(err)
				_, err = opts.Bot.SendMessage(context.Background(), params)
			}
		}

		err = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
	})

	opts.Cm.StaticCallback(MiningData, func(query telego.CallbackQuery) {
		if query.From.ID != 725757421 {
			_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "В разработке",
			})
			return
		}
		b, err := opts.BrakRepo.FindByUserID(query.From.ID, opts.M)
		qParams := &telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		}
		if err != nil {
			opts.Log.Sugar().Error(err)
			_ = opts.Bot.AnswerCallbackQuery(context.Background(), qParams.
				WithText(err.Error()),
			)
			return
		}
		params := &telego.SendPhotoParams{
			ChatID:    tu.ID(query.Message.GetChat().ID),
			ParseMode: telego.ModeHTML,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
			DisableNotification: true,
		}
		if b.Events.Mining == nil {
			photo, err := os.Open("resources/images/mining-cat.png")
			if err != nil {
				opts.Log.Sugar().Error(err)
				return
			}
			buyCallback := opts.Cm.DynamicCallback(callback.DynamicOpts{
				Label:    "КУПИТЬ ВСЕГО ЗА 5 000 000",
				CtxType:  callback.OneClick,
				OwnerIDs: []int64{query.From.ID},
				Time:     time.Minute * 45,
				Callback: func(query telego.CallbackQuery) {
					if b.Score < 5_000_000 {
						_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
							Text:            "У вас недостаточно средств для майнинг фермы.",
							CallbackQueryID: query.ID,
						})
						return
					}
					err = opts.BrakRepo.Update(bson.M{"_id": b.OID}, bson.M{
						"$inc": bson.M{
							"score": -5_000_000,
						},
						"$set": bson.M{
							"events.mining": &mining.Mining{},
						},
					})
					if err != nil {
						opts.Log.Sugar().Error(err)
						return
					}
					bPhoto, err := os.Open("resources/images/mining.jpg")
					_, err = opts.Bot.SendPhoto(context.Background(), &telego.SendPhotoParams{
						ChatID:              tu.ID(query.Message.GetChat().ID),
						Photo:               tu.File(bPhoto),
						Caption:             html.Bold("Поздравляем! Теперь вы владелец майнинг фермы и каждый час у вас есть шанс добыть биткоин который автоматически продаётся по выгодному курсу хинкалей!"),
						ParseMode:           telego.ModeHTML,
						DisableNotification: true,
					})
					if err != nil {
						opts.Log.Sugar().Error(err)
					}
				},
			})
			_, err = opts.Bot.SendPhoto(context.Background(), params.
				WithPhoto(tu.File(photo)).
				WithReplyMarkup(tu.InlineKeyboard(tu.InlineKeyboardRow(buyCallback.Inline()))).
				WithCaption(html.Bold("Есть пару лишних хинкалей, но для счастья твоей второй половинки всё равно не хватает? Купи брутальную майнинг ферму для своей семьи всего за 5 000 000 хинкалей и начинай майнить хинкали по крупному!\n")+"(Каждый час есть шанс добыть биткоин который автоматически продаётся по выгодному курсу хинкалей)"))
			err = photo.Close()
			if err != nil {
				opts.Log.Sugar().Error(err)
			}
		}
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
		err = opts.Bot.AnswerCallbackQuery(context.Background(), qParams)
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
	})
}
