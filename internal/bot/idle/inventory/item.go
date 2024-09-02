package inventory

import "famoria/internal/bot/idle/events"

type Item struct {
	Name         string
	CurrentLevel int
	MaxLevel     int
	Buffs        []events.Buff
}
