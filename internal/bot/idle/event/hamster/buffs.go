package hamster

import (
	"famoria/internal/bot/idle/event"
	"fmt"
)

//====== PlayPowerBuff ======

type PlayPowerBuff struct {
	Power uint64
}

func (b *PlayPowerBuff) Apply(base *event.Base) {
	base.BasePlayPower = b.Power
}

func (b *PlayPowerBuff) Type() event.GameType {
	return event.Hamster
}

func (b *PlayPowerBuff) Description() string {
	return fmt.Sprintf("+ %v к базовой силе тапа.", b.Power)
}

//====== PercentagePowerBuff ======

type PercentagePowerBuff struct {
	Percentage float64
}

func (b *PercentagePowerBuff) Apply(base *event.Base) {
	base.PercentagePower += b.Percentage
}

func (b *PercentagePowerBuff) Type() event.GameType {
	return event.Hamster
}

func (b *PercentagePowerBuff) Description() string {
	return fmt.Sprintf("+ %v%% силы тапа.", b.Percentage*100)
}

//====== PlayCountBuff ======

type PlayCountBuff struct {
	Count uint16
}

func (b *PlayCountBuff) Apply(base *event.Base) {
	base.MaxPlayCount += b.Count
}

func (b *PlayCountBuff) Type() event.GameType {
	return event.Hamster
}

func (b *PlayCountBuff) Description() string {
	return fmt.Sprintf("+ %v тапов в день.", b.Count)
}
