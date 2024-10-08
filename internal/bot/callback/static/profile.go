package static

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/event/anubis"
	"famoria/internal/bot/idle/event/casino"
	"famoria/internal/bot/idle/event/growkid"
	"famoria/internal/bot/idle/event/hamster"
	"famoria/internal/bot/idle/item"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
)

const (
	GrowKidData = "grow_kid"
	CasinoData  = "casino"
	HamsterData = "hamster"
	AnubisData  = "anubis"
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
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
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

		if response.IsWin {
			b.Score.Increase(response.Score)
		} else if response.Score != 0 {
			b.Score.Decrease(response.Score)
		}

		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$set": bson.M{
					"score":         b.Score,
					"events.casino": b.Events.Casino,
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error("Ошибка при обновлении счёта #casino (", response.Score, response.IsWin, ") пользователя ", query.From.ID, ":", err)
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
			_, err = opts.Bot.SendMessage(params)
			if err != nil {
				opts.Log.Sugar().Error(err)
			}
		} else {
			gif, err := os.Open(response.Path)
			if err == nil {
				_, err = opts.Bot.SendAnimation(&telego.SendAnimationParams{
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
				_, err = opts.Bot.SendMessage(params)
			}
		}

		if err != nil {
			opts.Log.Sugar().Error(err)
		}
		err = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
	})

	opts.Cm.StaticCallback(GrowKidData, func(query telego.CallbackQuery) {
		b, err := opts.BrakRepo.FindByUserID(query.From.ID, opts.M)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для кормления ребёнка необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}
		if b.BabyUserID == nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
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
		b.Score.Increase(response.Score)

		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$set": bson.M{
					"score":           b.Score,
					"events.grow_kid": b.Events.GrowKid,
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error("Ошибка при обновлении счёта #grow_kid (", response.Score, ") пользователя ", query.From.ID, ":", err)
			return
		}
		_, _ = opts.Bot.SendMessage(&telego.SendMessageParams{
			ChatID:    tu.ID(query.Message.GetChat().ID),
			ParseMode: telego.ModeHTML,
			Text:      response.Text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
		})
		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
	})

	opts.Cm.StaticCallback(HamsterData, func(query telego.CallbackQuery) {
		b, err := opts.BrakRepo.FindByUserID(query.From.ID, opts.M)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
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
		b.Score.Increase(response.Score)
		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$set": bson.M{
					"score":          b.Score,
					"events.hamster": b.Events.Hamster,
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error("Ошибка при обновлении счёта #hamster (", response.Score, ") пользователя ", query.From.ID, ":", err)
			return
		}

		err = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "Успешный тап по хомяку",
		})
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
	})

	opts.Cm.StaticCallback(AnubisData, func(query telego.CallbackQuery) {
		b, err := opts.BrakRepo.FindByUserID(query.From.ID, opts.M)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для игры в анубис вы должны быть в браке и иметь действующую подписку.",
				ShowAlert:       true,
			})
			return
		}
		response := b.Events.Anubis.Play(&anubis.PlayOpts{
			Log:      opts.Log,
			Bot:      opts.Bot,
			Query:    query,
			IsSub:    b.IsSub(),
			OldScore: b.Score,
		})
		if response == nil {
			return
		}

		if response.IsWin {
			b.Score.Increase(response.Score)
		} else if response.Score != 0 {
			b.Score.Decrease(response.Score)
		}

		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$set": bson.M{
					"score":         b.Score,
					"events.anubis": b.Events.Anubis,
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error("Ошибка при обновлении счёта #anubis (", response.Score, response.IsWin, ") пользователя ", query.From.ID, ":", err)
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
			_, err = opts.Bot.SendMessage(params)
			if err != nil {
				opts.Log.Sugar().Error(err)
			}
		} else {
			gif, err := os.Open(response.Path)
			if err == nil {
				_, err = opts.Bot.SendAnimation(&telego.SendAnimationParams{
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
				_, err = opts.Bot.SendMessage(params)
			}
		}

		err = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
	})
}
