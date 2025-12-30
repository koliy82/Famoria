package growkid

import (
	"context"
	"famoria/internal/bot/idle/event"
	"famoria/internal/pkg/date"
	"famoria/internal/pkg/html"
	"fmt"
	"math/rand"
	"time"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type GrowKid struct {
	event.Base `bson:"base"`
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
	Score int64
	Text  string
}

func (g *GrowKid) Play(opts *PlayOpts) *PlayResponse {
	if !date.HasUpdated(g.LastPlay) {
		g.PlayCount = g.MaxPlayCount
		g.LastPlay = time.Now()
	}

	if g.PlayCount == 0 {
		_ = opts.Bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
			CallbackQueryID: opts.Query.ID,
			Text:            "Вы сегодня уже кормили ребёнка.",
			ShowAlert:       true,
		})
		return nil
	}

	score := int64((float64(rand.Int63n(50))+float64(g.BasePlayPower))*g.PercentagePower) + 1
	g.PlayCount--
	return &PlayResponse{
		Score: score,
		Text:  fmt.Sprintf("%s покормил своего ребёнка и получил %d хинкалей!", html.UserMention(&opts.Query.From), score),
	}
}
