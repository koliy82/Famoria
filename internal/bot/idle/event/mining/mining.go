package mining

import (
	"famoria/internal/bot/idle/event"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type Mining struct {
	event.Base `bson:"base"`
}

func (c *Mining) DefaultStats() {
	if c == nil {
		return
	}
	c.Base.MaxPlayCount = 1
	c.Base.PercentagePower = 1.0
	c.Base.BasePlayPower = 1000
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
