package hamster

import (
	"famoria/internal/bot/idle/events"
	"famoria/internal/pkg/date"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"time"
)

type Hamster struct {
	events.Base `bson:"base"`
}

func (h *Hamster) DefaultStats() {
	h.Base.MaxPlayCount = 20
	h.Base.BasePlayPower = 1
	h.Base.PercentagePower = 1.0
}

type PlayOpts struct {
	Log   *zap.Logger
	Bot   *telego.Bot
	Query telego.CallbackQuery
}

type PlayResponse struct {
	Score uint64
}

func (h *Hamster) Play(opts *PlayOpts) *PlayResponse {
	if !date.HasUpdated(h.LastPlay) {
		h.PlayCount = h.MaxPlayCount
		h.LastPlay = time.Now()
	}

	if h.PlayCount == 0 {
		_ = opts.Bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: opts.Query.ID,
			Text:            "Хомяк устал, он разрешит себя тапать завтра.",
			ShowAlert:       true,
		})
		return nil
	}

	score := uint64(float64(h.BasePlayPower) * h.PercentagePower)
	h.PlayCount--

	return &PlayResponse{Score: score}
}
