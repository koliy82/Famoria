package mining

import (
	"famoria/internal/bot/idle/event"
	"fmt"
	"math/rand"
)

type Mining struct {
	event.Base `bson:"base"`
}

func (m *Mining) DefaultStats() {
	if m == nil {
		return
	}
	m.Base.MaxPlayCount = 1
	m.Base.PercentagePower = 1.0
	m.Base.BasePlayPower = 500_000
}

type PlayResponse struct {
	Score int64
	Text  string
	IsWin bool
}

func (m *Mining) Play() *PlayResponse {
	chance := rand.Intn(101) + m.Luck
	score := int64(float64(m.BasePlayPower)*m.PercentagePower) + 1
	switch {
	case chance <= 4:
		score = score / 5
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("сгорела ферма, нужно тех обсл. -%d", score),
			IsWin: false,
		}
	case chance <= 50:
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("Нашёл хилькоин, продано за %d хинкалей!", score),
			IsWin: false,
		}
	case chance >= 90:
		return &PlayResponse{
			Score: score,
			Text:  fmt.Sprintf("Сегодня даёт, сорвал %d хинкалей!", score),
			IsWin: true,
		}
	default:
		return &PlayResponse{
			Score: 0,
			Text:  fmt.Sprintf("Ферма без успехов."),
			IsWin: false,
		}
	}
}
