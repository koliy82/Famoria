package anubis

import (
	"context"
	"famoria/internal/bot/idle/event"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/date"
	"famoria/internal/pkg/html"
	"fmt"
	"math/rand"
	"time"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type Anubis struct {
	event.Base `bson:"base"`
}

func (a *Anubis) DefaultStats() {
	a.Base.MaxPlayCount = 3
	a.Base.PercentagePower = 1.0
	a.Base.BasePlayPower = 1000
}

type PlayOpts struct {
	Log      *zap.Logger
	Bot      *telego.Bot
	Query    telego.CallbackQuery
	IsSub    bool
	OldScore *common.Score
}

type PlayResponse struct {
	Score int64
	Text  string
	IsWin bool
	Path  string
}

func (a *Anubis) Play(opts *PlayOpts) *PlayResponse {
	if !date.HasUpdated(a.LastPlay) {
		a.PlayCount = a.MaxPlayCount
		a.LastPlay = time.Now()
	}

	if a.PlayCount == 0 {
		_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
			CallbackQueryID: opts.Query.ID,
			Text:            "Сегодня вы уже прошли испытание Анубиса.",
			ShowAlert:       true,
		})
		return nil
	}

	a.PlayCount--

	if !opts.IsSub {
		score := int64(rand.Intn(100) + 1)

		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s ты пытался обмануть Анубиса? У тебя нет подписки, анубис забрал у тебя %v хинкалей.", html.UserMention(&opts.Query.From), score),
			IsWin: false,
		}
	}

	chance := rand.Intn(101) + a.Luck
	score := int64(float64(uint64(rand.Int31n(200))+a.BasePlayPower)*a.PercentagePower) + 1
	switch {
	case chance == 1:
		score *= 2
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s не прошёл испытание Анубиса и потерял %d хинкалей.", html.UserMention(&opts.Query.From), score),
			IsWin: false,
		}
	case chance <= 15:
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s попал в ловушку, поставленную Анубисом, и проиграл %d хинкалей.", html.UserMention(&opts.Query.From), score),
			IsWin: false,
		}
	case chance <= 35:
		return &PlayResponse{
			Score: 0,
			Text:  fmt.Sprintf("%s встретился с Анубисом, но ему удалось избежать потерь.", html.UserMention(&opts.Query.From)),
			IsWin: false,
		}
	case chance <= 70:
		score *= 5
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s прошёл испытание и заработал %d хинкалей.", html.UserMention(&opts.Query.From), score),
			IsWin: true,
		}
	case chance <= 85:
		score *= 10
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s победил Анубиса и обнаружил скрытый клад в %d хинкалей!", html.UserMention(&opts.Query.From), score),
			IsWin: true,
		}
	case chance <= 99:
		score *= 50
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("Анубис сегодня даёт, %s сорвал огромный куш в %d хинкалей!", html.UserMention(&opts.Query.From), score),
			IsWin: true,
		}
	case chance <= 100:
		oldShow := *opts.OldScore
		opts.OldScore.Multiply(1.2)
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("Анубис использовал древнюю магию и умножил счёт %s на 20%%.\nБаланс %s -> %s. \n(Вы также обнаружили скрытый сундук и нашли в нём %v хинкаль)", html.UserMention(&opts.Query.From), oldShow.GetFormattedScore(), opts.OldScore.GetFormattedScore(), score),
			IsWin: true,
			Path:  "resources/gifs/bigwin.gif.mp4",
		}
	default:
		return &PlayResponse{
			Score: 0,
			Text:  fmt.Sprintf("%s пытался пройти испытание Анубиса, но остался в нуле.", html.UserMention(&opts.Query.From)),
			IsWin: false,
		}
	}
}
