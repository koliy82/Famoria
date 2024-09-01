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

		if !date.HasUpdated(b.Casino.LastPlay) {
			b.Casino.PlayCount = b.Casino.MaxPlayCount
			b.Casino.LastPlay = time.Now()
		}

		if b.Casino.PlayCount == 0 {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Сегодня вы уже играли в казино.",
				ShowAlert:       true,
			})
			return
		}

		score := uint64(rand.Int31n(500))
		chance := rand.Intn(100)
		text := ""
		switch {
		case chance <= 40:
			text = fmt.Sprintf("%s выйграл в казино %d хинкалей!", html.UserMention(&query.From), score)
			b.Score.Increase(score)
		case chance <= 70:
			text = fmt.Sprintf("%s заигрался в казино и влез в кредит на %d хинкалей!", html.UserMention(&query.From), score)
			b.Score.Decrease(score)
		case chance <= 75:
			score = score * 2
			text = fmt.Sprintf("%s выйграл в казино %d хинкалей, весьма неплохо!", html.UserMention(&query.From), score)
			b.Score.Increase(score)
		case chance == 76:
			score = score * 6
			b.Score.Increase(score * 5)
			text = fmt.Sprintf("%s сорвал куш на %d хинкалей.", html.UserMention(&query.From), score)
		case chance == 77:
			score = score * 3
			b.Score.Decrease(score * 2)
			text = fmt.Sprintf("%s сегодня не везёт, он проиграл %d хинкалей.", html.UserMention(&query.From), score)
		default:
			text = fmt.Sprintf("%s играл сегодня в казино, но остался в нуле.", html.UserMention(&query.From))
		}

		b.Casino.PlayCount -= 1
		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$set": bson.M{
					"score":  b.Score,
					"casino": b.Casino,
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

		if !date.HasUpdated(b.GrowKid.LastPlay) {
			b.GrowKid.PlayCount = b.GrowKid.MaxPlayCount
			b.GrowKid.LastPlay = time.Now()
		}

		if b.GrowKid.PlayCount == 0 {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Вы сегодня уже кормили ребёнка.",
				ShowAlert:       true,
			})
			return
		}

		score := uint64(rand.Int31n(50) + 20)
		b.Score.Increase(score)
		b.GrowKid.PlayCount -= 1

		err = opts.BrakRepo.Update(
			bson.M{"_id": b.OID},
			bson.M{
				"$set": bson.M{
					"score":    b.Score,
					"grow_kid": b.GrowKid,
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
				Text:            "Для тапа хомяка необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}

		if !date.HasUpdated(b.Hamster.LastPlay) {
			b.Score.Increase(b.Hamster.PlayPower)
			b.Hamster.PlayCount = b.Hamster.MaxPlayCount - 1
			b.Hamster.LastPlay = time.Now()
			err = opts.BrakRepo.Update(
				bson.M{"_id": b.OID},
				bson.M{
					"$set": bson.M{
						"score":   b.Score,
						"hamster": b.Hamster,
					},
				},
			)
		} else if b.Hamster.PlayCount == 0 {
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Хомяк устал, он разрешит себя тапать завтра.",
				ShowAlert:       true,
			})
			return
		} else {
			b.Score.Increase(b.Hamster.PlayPower)
			b.Hamster.PlayCount -= 1
			err = opts.BrakRepo.Update(
				bson.M{"_id": b.OID},
				bson.M{
					"$set": bson.M{
						"score":   b.Score,
						"hamster": b.Hamster,
					},
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
