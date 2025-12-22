package events

import (
	"famoria/internal/bot/idle/event"
	"famoria/internal/bot/idle/event/anubis"
	"famoria/internal/bot/idle/event/casino"
	"famoria/internal/bot/idle/event/growkid"
	"famoria/internal/bot/idle/event/hamster"
	"famoria/internal/bot/idle/event/mining"
	"time"
)

type Events struct {
	Casino  *casino.Casino   `bson:"casino"`
	Hamster *hamster.Hamster `bson:"hamster"`
	GrowKid *growkid.GrowKid `bson:"grow_kid"`
	Anubis  *anubis.Anubis   `bson:"anubis"`
	Mining  *mining.Mining   `bson:"mining"`
	Shop    *event.Base      `bson:"-"`
}

func (e *Events) DefaultStats() {
	e.Casino.DefaultStats()
	e.Hamster.DefaultStats()
	e.GrowKid.DefaultStats()
	e.Anubis.DefaultStats()
	e.Mining.DefaultStats()
}

func New() *Events {
	return &Events{
		Hamster: &hamster.Hamster{
			Base: event.Base{
				LastPlay:  time.Time{},
				PlayCount: 0,
			},
		},
		Casino: &casino.Casino{
			Base: event.Base{
				LastPlay:  time.Time{},
				PlayCount: 0,
			},
		},
		GrowKid: &growkid.GrowKid{
			Base: event.Base{
				LastPlay:  time.Time{},
				PlayCount: 0,
			},
		},
		Anubis: &anubis.Anubis{
			Base: event.Base{
				LastPlay:  time.Time{},
				PlayCount: 0,
			},
		},
	}
}
