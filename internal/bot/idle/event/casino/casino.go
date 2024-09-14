package casino

import (
	"famoria/internal/bot/idle/event"
	"famoria/internal/pkg/date"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

type Casino struct {
	event.Base `bson:"base"`
}

func (c *Casino) DefaultStats() {
	c.Base.MaxPlayCount = 1
	c.Base.PercentagePower = 1.0
	c.Base.BasePlayPower = 300
}

type PlayOpts struct {
	Log   *zap.Logger
	Bot   *telego.Bot
	Query telego.CallbackQuery
}

type PlayResponse struct {
	Score uint64
	Text  string
	IsWin bool
	Path  string
}

func (c *Casino) Play(opts *PlayOpts) *PlayResponse {
	if !date.HasUpdated(c.LastPlay) {
		c.PlayCount = c.MaxPlayCount
		c.LastPlay = time.Now()
	}

	if c.PlayCount == 0 {
		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: opts.Query.ID,
			Text:            "Сегодня вы уже играли в казино.",
			ShowAlert:       true,
		})
		return nil
	}

	chance := rand.Intn(100) + c.Luck
	score := uint64(float64(uint64(rand.Int31n(100))+c.BasePlayPower)*c.PercentagePower) + 1
	c.PlayCount--
	chance = 100
	switch {
	case chance == 1:
		score *= 3
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s сегодня не везёт, он проиграл %d хинкалей.", html.UserMention(&opts.Query.From), score),
			IsWin: false,
		}
	case chance <= 35:
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s заигрался в казино и влез в кредит на %d хинкалей!", html.UserMention(&opts.Query.From), score),
			IsWin: false,
		}
	case chance <= 55:
		return &PlayResponse{
			Score: 0,
			Text:  fmt.Sprintf("%s играл сегодня в казино, но остался в нуле.", html.UserMention(&opts.Query.From)),
			IsWin: false,
		}
	case chance <= 70:
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s выйграл в казино %d хинкалей!", html.UserMention(&opts.Query.From), score),
			IsWin: true,
		}
	case chance <= 85:
		score *= 2
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s выйграл в казино %d хинкалей, весьма неплохо!", html.UserMention(&opts.Query.From), score),
			IsWin: true,
		}
	case chance <= 100:
		score *= 6
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("%s сорвал куш на %d хинкалей.", html.UserMention(&opts.Query.From), score),
			IsWin: true,
			Path:  "resources/gifs/papich_win.gif.mp4",
		}
	default:
		return &PlayResponse{
			Score: 0,
			Text:  fmt.Sprintf("%s играл сегодня в казино, но остался в нуле.", html.UserMention(&opts.Query.From)),
			IsWin: false,
		}
	}
}
