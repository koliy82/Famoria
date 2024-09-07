package subscribe

import "famoria/internal/bot/idle/events"

// ====== AnubisBuff ======

type AnubisBuff struct{}

func (b *AnubisBuff) Apply(*events.Base) {}

func (b *AnubisBuff) Type() events.GameType {
	return events.Subscribe
}

func (b *AnubisBuff) Description() string {
	return "+ Доступ к игре в Анубис."
}
