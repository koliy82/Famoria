package anubis

import "famoria/internal/bot/idle/event"

// ====== AccessBuff ======

type AccessBuff struct{}

func (b *AccessBuff) Apply(*event.Base) {}

func (b *AccessBuff) Type() event.GameType {
	return event.Subscribe
}

func (b *AccessBuff) Description() string {
	return "+ Доступ к игре в Анубис."
}
