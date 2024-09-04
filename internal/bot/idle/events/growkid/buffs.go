package growkid

import (
	"famoria/internal/bot/idle/events"
	"famoria/internal/pkg/html"
	"fmt"
	"strconv"
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
	return fmt.Sprintf("+ %s к базовой кормёжке.", html.Bold(strconv.FormatUint(b.Power, 10)))
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
	return fmt.Sprintf("+ %s%% силы кормёжки.", html.Bold(strconv.FormatFloat(b.Percentage*100, 'f', 2, 64)))
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
	return fmt.Sprintf("+ %s кормёжок в день.", html.Bold(strconv.Itoa(int(b.Count))))
}
