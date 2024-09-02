package hamster

import (
	"famoria/internal/bot/idle/events"
)

type Hamster struct {
	events.Base `bson:"base"`
}
