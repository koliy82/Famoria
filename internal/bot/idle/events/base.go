package events

import (
	"time"
)

type GameType int

const (
	Hamster GameType = iota
	Casino
	GrowKid
)

type Base struct {
	LastPlay        time.Time `bson:"last_play"`
	PlayCount       uint16    `bson:"play_count"`
	MaxPlayCount    uint16    `bson:"max_play_count"`
	BasePlayPower   uint64    `bson:"play_power"`
	PercentagePower uint32    `bson:"percentage_power"`
}

type Buff interface {
	Type() GameType
	Apply(*Base)
}
