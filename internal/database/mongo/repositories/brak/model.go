package brak

import (
	"famoria/internal/bot/idle/events"
	"famoria/internal/bot/idle/events/casino"
	"famoria/internal/bot/idle/events/growkid"
	"famoria/internal/bot/idle/events/hamster"
	"famoria/internal/bot/idle/inventory"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/plural"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Brak struct {
	OID            primitive.ObjectID   `bson:"_id"`
	FirstUserID    int64                `bson:"first_user_id"`
	SecondUserID   int64                `bson:"second_user_id"`
	ChatID         int64                `bson:"chat_id,omitempty"`
	CreateDate     time.Time            `bson:"create_date"`
	BabyUserID     *int64               `bson:"baby_user_id"`
	BabyCreateDate *time.Time           `bson:"baby_create_date"`
	Score          common.Score         `bson:"score"`
	Inventory      *inventory.Inventory `bson:"inventory"`
	Casino         *casino.Casino       `bson:"casino"`
	Hamster        *hamster.Hamster     `bson:"hamster"`
	GrowKid        *growkid.GrowKid     `bson:"grow_kid"`

	//LastCasinoPlay    time.Time          `bson:"last_casino_play"`
	//LastGrowKid       time.Time          `bson:"last_grow_kid"`
	//LastHamsterUpdate time.Time          `bson:"last_hamster_update"`
	//TapCount          int                `bson:"tap_count"`
}

func (b Brak) ApplyBuffs() {
	for _, i := range b.Inventory.Items {
		for _, buff := range i.Buffs {
			switch buff.Type() {
			case events.Hamster:
				buff.Apply(&b.Hamster.Base)
			case events.Casino:
				buff.Apply(&b.Casino.Base)
			case events.GrowKid:
				buff.Apply(&b.GrowKid.Base)
			}
		}
	}
}

type UsersBrak struct {
	Brak   *Brak
	First  *user.User
	Second *user.User
	Baby   *user.User
}

// PartnerID returns the partner's ID by the user's ID
func (b Brak) PartnerID(userID int64) int64 {
	if b.FirstUserID == userID {
		return b.SecondUserID
	}
	return b.FirstUserID
}

// Duration returns the duration of the relationship
func (b Brak) Duration() string {
	duration := time.Now().Sub(b.CreateDate)
	hours := int(duration.Hours())

	if hours < 1 {
		return "молодожены"
	}

	days := hours / 24
	months := days / 30
	years := days / 365

	if years > 0 {
		return fmt.Sprintf("%d %s", years, plural.Declension(years, "год", "года", "лет"))
	}

	if months > 0 {
		return fmt.Sprintf("%d %s", months, plural.Declension(months, "месяц", "месяца", "месяцев"))
	}

	if days > 0 {
		return fmt.Sprintf("%d %s", days, plural.Declension(days, "день", "дня", "дней"))
	}

	return fmt.Sprintf("%d %s", hours, plural.Declension(hours, "час", "часа", "часов"))
}

func (b Brak) DurationKid() string {
	if b.BabyCreateDate == nil {
		return ""
	}
	duration := time.Now().Sub(*b.BabyCreateDate)
	seconds := int(duration.Seconds())

	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	months := days / 30
	years := days / 365

	switch {
	case years > 0:
		return fmt.Sprintf("%d %s", years, plural.Declension(years, "год", "года", "лет"))
	case months > 0:
		return fmt.Sprintf("%d %s", months, plural.Declension(months, "месяц", "месяца", "месяцев"))
	case days > 0:
		return fmt.Sprintf("%d %s", days, plural.Declension(days, "день", "дня", "дней"))
	case hours > 0:
		return fmt.Sprintf("%d %s", hours, plural.Declension(hours, "час", "часа", "часов"))
	case minutes > 0:
		return fmt.Sprintf("%d %s", minutes, plural.Declension(minutes, "минута", "минуты", "минут"))
	default:
		return fmt.Sprintf("%d %s", seconds, plural.Declension(seconds, "секунда", "секунды", "секунд"))
	}
}
