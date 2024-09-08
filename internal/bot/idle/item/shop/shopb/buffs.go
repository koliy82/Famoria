package shopb

import (
	"famoria/internal/bot/idle/event"
	"fmt"
)

//====== SaleBuff ======

type SaleBuff struct {
	Percentage float64
}

func (b *SaleBuff) Apply(base *event.Base) {
	base.Sale += b.Percentage
}

func (b *SaleBuff) Type() event.GameType {
	return event.Shop
}

func (b *SaleBuff) Description() string {
	return fmt.Sprintf("+ %v%% скидка в потайной лавке.", b.Percentage*100)
}
