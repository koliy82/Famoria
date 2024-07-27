package static

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
	"math/rand"
)

const (
	GrowKidData = "grow_kid"
	CasinoData  = "casino"
)

type Opts struct {
	fx.In
	Log   *zap.Logger
	Braks brak.Repository
	Users user.Repository
	Cm    *callback.CallbacksManager
	Bot   *telego.Bot
}

func ProfileCallbacks(opts Opts) {
	opts.Cm.StaticCallback(CasinoData, func(query telego.CallbackQuery) {
		b, err := opts.Braks.FindByUserID(query.From.ID)
		if err != nil {
			opts.Log.Sugar().Error(err)
			_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Для использования казино необходимо жениться.",
				ShowAlert:       true,
			})
			return
		}
		score := rand.Intn(200) - 100
		err = opts.Braks.UpdateScore(b.OID, score)
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
			text = fmt.Sprintf("You win %d!", score)
		case score < 0:
			text = fmt.Sprintf("You lose %d!", score)
		default:
			text = "You don't win or lose."
		}
		_, _ = opts.Bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(query.Message.GetChat().ID),
			Text:   text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: query.Message.GetMessageID(),
			},
		})
		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		})
	})

	opts.Cm.StaticCallback(GrowKidData, func(query telego.CallbackQuery) {

	})
}
