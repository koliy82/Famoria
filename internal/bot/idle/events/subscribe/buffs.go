package subscribe

import (
	"famoria/internal/bot/idle/events"
	"fmt"
)

//====== SubBuff ======

type SubBuff struct {
	percentage float64
}

func (b *SubBuff) Apply(base *events.Base) {
	base.PercentagePower += b.percentage
}

func (b *SubBuff) Type() events.GameType {
	return events.Subscribe
}

func (b *SubBuff) Description() string {
	return fmt.Sprintf("+ %v%% ко всему.", b.percentage)
}
