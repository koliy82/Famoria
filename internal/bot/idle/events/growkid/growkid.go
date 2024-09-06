package growkid

import (
	"famoria/internal/bot/idle/events"
	"famoria/internal/pkg/date"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

type GrowKid struct {
	events.Base `bson:"base"`
}

func (g *GrowKid) DefaultStats() {
	g.Base.MaxPlayCount = 1
	g.Base.PercentagePower = 1.0
	g.Base.BasePlayPower = 50
}

type PlayOpts struct {
	Log   *zap.Logger
	Bot   *telego.Bot
	Query telego.CallbackQuery
}

type PlayResponse struct {
	Score uint64
	Text  string
}

func (g *GrowKid) Play(opts *PlayOpts) *PlayResponse {
	if !date.HasUpdated(g.LastPlay) {
		g.PlayCount = g.MaxPlayCount
		g.LastPlay = time.Now()
	}

	if g.PlayCount == 0 {
		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: opts.Query.ID,
			Text:            "Вы сегодня уже кормили ребёнка.",
			ShowAlert:       true,
		})
		return nil
	}

	score := uint64(float64(rand.Int63n(int64(g.BasePlayPower)))*g.PercentagePower) + 1
	g.PlayCount--
	return &PlayResponse{
		Score: score,
		Text:  fmt.Sprintf("%s покормил своего ребёнка и получил %d хинкалей!", html.UserMention(&opts.Query.From), score),
	}
}
