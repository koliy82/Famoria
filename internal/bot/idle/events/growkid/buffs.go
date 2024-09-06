package growkid

import (
	"famoria/internal/bot/idle/events"
	"fmt"
)

//====== PlayPowerBuff ======

type PlayPowerBuff struct {
	Power uint64
}

func (b *PlayPowerBuff) Apply(base *events.Base) {
	base.BasePlayPower = b.Power
}

func (b *PlayPowerBuff) Type() events.GameType {
	return events.GrowKid
}

func (b *PlayPowerBuff) Description() string {
	return fmt.Sprintf("+ %v к базовой кормёжке.", b.Power)
}

//====== PercentagePowerBuff ======

type PercentagePowerBuff struct {
	Percentage float64
}

func (b *PercentagePowerBuff) Apply(base *events.Base) {
	base.PercentagePower += b.Percentage
}

func (b *PercentagePowerBuff) Type() events.GameType {
	return events.GrowKid
}

func (b *PercentagePowerBuff) Description() string {
	return fmt.Sprintf("+ %v%% силы кормёжки.", b.Percentage*100)
}

//====== PlayCountBuff ======

type PlayCountBuff struct {
	Count uint16
}

func (b *PlayCountBuff) Apply(base *events.Base) {
	base.MaxPlayCount += b.Count
}

func (b *PlayCountBuff) Type() events.GameType {
	return events.GrowKid
}

func (b *PlayCountBuff) Description() string {
	return fmt.Sprintf("+ %v кормёжок в день.", b.Count)
}
