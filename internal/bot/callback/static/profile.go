package static

import (
	"famoria/internal/bot/callback"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/date"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

const (
	GrowKidData = "grow_kid"
	CasinoData  = "casino"
	HamsterData = "hamster"
)

type Opts struct {
	fx.In
	Log      *zap.Logger
	BrakRepo brak.Repository
	UserRepo user.Repository
	Cm       *callback.CallbacksManager
	Bot      *telego.Bot
}

func ProfileCallbacks(opts Opts) {
	opts.Cm.StaticCallback(CasinoData, func(query telego.CallbackQuery) {
		b, err := opts.BrakRepo.FindByUserID(query.From.ID)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для использования казино необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}

		if date.HasUpdated(b.LastCasinoPlay) {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Играть в казино можно раз в сутки.",
				ShowAlert:       true,
			})
			return
		}

		score := rand.Int63n(500) - 300
		b.Score.IncreaseScore(score)

		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$set": bson.M{
					"score":            b.Score,
					"last_casino_play": time.Now(),
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error(err)
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Ошибка при обновлении счёта.",
				ShowAlert:       true,
			})
			return
		}
		text := ""
		switch {
		case score > 0:
			text = fmt.Sprintf("%s выйграл в казино %d хинкалей!", html.UserMention(&query.From), score)
		case score < 0:
			text = fmt.Sprintf("%s заигрался в казино и влез в кредит на %d хинкалей!", html.UserMention(&query.From), score)
		default:
			text = "%s играл сегодня в казино, но остался в нуле."
		}
		_, err = opts.Bot.SendMessage(&telego.SendMessageParams{
			ChatID:    tu.ID(query.Message.GetChat().ID),
			ParseMode: telego.ModeHTML,
			Text:      text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
		})
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
		b, err := opts.BrakRepo.FindByUserID(query.From.ID)
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

		if date.HasUpdated(b.LastGrowKid) {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Кормить ребёнка можно раз в сутки.",
				ShowAlert:       true,
			})
			return
		}

		score := rand.Int63n(30) + 20
		b.Score.IncreaseScore(score)

		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$set": bson.M{
					"score":         b.Score,
					"last_grow_kid": time.Now(),
				},
			},
		)
		if err != nil {
			opts.Log.Sugar().Error(err)
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Ошибка при обновлении счёта.",
				ShowAlert:       true,
			})
			return
		}
		text := ""
		switch {
		case score > 0:
			text = fmt.Sprintf("%s покормил своего ребёнка и получил от жены %d хинкалей!", html.UserMention(&query.From), score)
		case score < 0:
			text = fmt.Sprintf("You lose %d!", score)
		default:
			text = "You don't win or lose."
		}
		_, _ = opts.Bot.SendMessage(&telego.SendMessageParams{
			ChatID:    tu.ID(query.Message.GetChat().ID),
			ParseMode: telego.ModeHTML,
			Text:      text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
		})
		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
	})

	opts.Cm.StaticCallback(HamsterData, func(query telego.CallbackQuery) {
		b, err := opts.BrakRepo.FindByUserID(query.From.ID)
		if err != nil {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для использования казино необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}

		if !date.HasUpdated(b.LastHamsterUpdate) {
			b.Score.IncreaseScore(1)
			err = opts.BrakRepo.Update(
				bson.M{"_id": b.OID},
				bson.M{
					"$set": bson.M{
						"score":               b.Score,
						"tap_count":           49,
						"last_hamster_update": time.Now(),
					},
				},
			)
		} else if b.TapCount == 0 {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Хомяк устал, он разрешит себя тапать завтра.",
				ShowAlert:       true,
			})
			return
		} else {
			b.Score.IncreaseScore(1)
			err = opts.BrakRepo.Update(
				bson.M{"_id": b.OID},
				bson.M{
					"$set": bson.M{"score": b.Score},
					"$inc": bson.M{"tap_count": -1},
				},
			)
		}

		if err != nil {
			opts.Log.Sugar().Error(err)
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Ошибка при обновлении счёта.",
				ShowAlert:       true,
			})
			return
		}

		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "Успешный тап по хомяку",
		})
		return

	})
}
