package casino

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
	return events.Casino
}

func (b *PlayPowerBuff) Description() string {
	return fmt.Sprintf("+ %s к базовому выйгрышу.", html.Bold(strconv.FormatUint(b.Power, 10)))
}

//====== PercentagePowerBuff ======

type PercentagePowerBuff struct {
	Percentage float64
}

func (b *PercentagePowerBuff) Apply(base *events.Base) {
	base.PercentagePower += b.Percentage
}

func (b *PercentagePowerBuff) Type() events.GameType {
	return events.Casino
}

func (b *PercentagePowerBuff) Description() string {
	return fmt.Sprintf("+ %v%% выйгрыша.", b.Percentage)
}

//====== PlayCountBuff ======

type PlayCountBuff struct {
	Count uint16
}

func (b *PlayCountBuff) Apply(base *events.Base) {
	base.MaxPlayCount += b.Count
}

func (b *PlayCountBuff) Type() events.GameType {
	return events.Casino
}

func (b *PlayCountBuff) Description() string {
	return fmt.Sprintf("+ %s круток в день.", html.Bold(strconv.Itoa(int(b.Count))))
}

//====== LuckBuff ======

type LuckBuff struct {
	Luck int
}

func (b *LuckBuff) Apply(base *events.Base) {
	base.Luck += b.Luck
}

func (b *LuckBuff) Type() events.GameType {
	return events.Casino
}

func (b *LuckBuff) Description() string {
	return fmt.Sprintf("+ %s%% удачи.", html.Bold(strconv.Itoa(b.Luck)))
}
