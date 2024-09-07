package shopb

import (
	"famoria/internal/bot/idle/events"
	"fmt"
)

//====== SaleBuff ======

type SaleBuff struct {
	Percentage float64
}

func (b *SaleBuff) Apply(base *events.Base) {
	base.Sale += b.Percentage
}

func (b *SaleBuff) Type() events.GameType {
	return events.Shop
}

func (b *SaleBuff) Description() string {
	return fmt.Sprintf("+ %v%% скидка в потайной лавке.", b.Percentage*100)
}
