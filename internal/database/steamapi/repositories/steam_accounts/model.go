package steam_accounts

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SteamAccount struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	TelegramID   int64              `bson:"telegram_id"`
	Username     *string            `bson:"username,omitempty"`
	Password     *string            `bson:"password,omitempty"`
	RefreshToken *string            `bson:"refresh_token,omitempty"`
	GameIDs      []uint32           `bson:"game_ids,omitempty"`
	IsFarming    bool               `bson:"is_farming"`
	PersonaState PersonaState       `bson:"persona_state"`
}

func (a *SteamAccount) Name() string {
	if a.Username == nil {
		return "?"
	}
	return *a.Username
}

func (a *SteamAccount) Games() string {
	return strings.Trim(strings.Replace(fmt.Sprint(a.GameIDs), " ", ", ", -1), "[]")
}

type PersonaState int

const (
	Offline PersonaState = iota
	Online
	Busy
	Away
	Snooze
	LookingToTrade
	LookingToPlay
	Invisible
)

func (s PersonaState) String() string {
	return [...]string{"Офлайн", "В сети", "Занят", "Нет на месте", "Спит", "Ищет трейды", "Ищет игры", "Невидимка"}[s]
}

func AvailableStates() []PersonaState {
	return []PersonaState{Online, Busy, Away, Snooze, LookingToTrade, LookingToPlay, Invisible}
}
