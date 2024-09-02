package hamster

import "famoria/internal/bot/idle/events"

//====== BaseHamsterBuff ======

type BaseHamsterBuff struct {
	power uint64
}

func (b *BaseHamsterBuff) Apply(base *events.Base) {
	base.BasePlayPower = b.power
}

func (b *BaseHamsterBuff) Type() events.GameType {
	return events.Hamster
}

//====== PercentageHamsterBuff ======

type PercentageHamsterBuff struct {
	percentage uint32
}

func (b *PercentageHamsterBuff) Apply(base *events.Base) {
	base.PercentagePower += b.percentage
}

func (b *PercentageHamsterBuff) Type() events.GameType {
	return events.Hamster
}

//====== ... ======
