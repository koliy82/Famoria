package events

import "time"

type Base struct {
	LastPlay     time.Time `bson:"last_play"`
	PlayCount    uint16    `bson:"play_count"`
	MaxPlayCount uint16    `bson:"max_play_count"`
	PlayPower    uint64    `bson:"play_power"`
}
