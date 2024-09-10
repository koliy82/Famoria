package event

import (
	"time"
)

type GameType int

const (
	Hamster GameType = iota
	Casino
	GrowKid
	Shop
	Subscribe
)

type Base struct {
	LastPlay        time.Time `bson:"last_play"`
	PlayCount       uint16    `bson:"play_count"`
	MaxPlayCount    uint16    `bson:"-"`
	BasePlayPower   uint64    `bson:"-"`
	PercentagePower float64   `bson:"-"`
	Luck            int       `bson:"-"`
	Sale            float64   `bson:"-"`
}

type Buff interface {
	Type() GameType
	Apply(*Base)
	Description() string
}
