package brak

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go_tg_bot/internal/database/mongo/repositories/user"
	"go_tg_bot/internal/utils/date"
	"time"
)

type Brak struct {
	OID               primitive.ObjectID `bson:"_id"`
	FirstUserID       int64              `bson:"first_user_id"`
	SecondUserID      int64              `bson:"second_user_id"`
	ChatID            int64              `bson:"chat_id"`
	CreateDate        time.Time          `bson:"create_date"`
	BabyUserID        *int64             `bson:"baby_user_id"`
	BabyCreateDate    *time.Time         `bson:"baby_create_date"`
	Score             int64              `bson:"score"`
	LastCasinoPlay    time.Time          `bson:"last_casino_play"`
	LastGrowKid       time.Time          `bson:"last_grow_kid"`
	LastHamsterUpdate time.Time          `bson:"last_hamster_update"`
	TapCount          int                `bson:"tap_count"`
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
		return fmt.Sprintf("%d %s", years, date.Declension(years, "год", "года", "лет"))
	}

	if months > 0 {
		return fmt.Sprintf("%d %s", months, date.Declension(months, "месяц", "месяца", "месяцев"))
	}

	if days > 0 {
		return fmt.Sprintf("%d %s", days, date.Declension(days, "день", "дня", "дней"))
	}

	return fmt.Sprintf("%d %s", hours, date.Declension(hours, "час", "часа", "часов"))
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
		return fmt.Sprintf("%d %s", years, date.Declension(years, "год", "года", "лет"))
	case months > 0:
		return fmt.Sprintf("%d %s", months, date.Declension(months, "месяц", "месяца", "месяцев"))
	case days > 0:
		return fmt.Sprintf("%d %s", days, date.Declension(days, "день", "дня", "дней"))
	case hours > 0:
		return fmt.Sprintf("%d %s", hours, date.Declension(hours, "час", "часа", "часов"))
	case minutes > 0:
		return fmt.Sprintf("%d %s", minutes, date.Declension(minutes, "минута", "минуты", "минут"))
	default:
		return fmt.Sprintf("%d %s", seconds, date.Declension(seconds, "секунда", "секунды", "секунд"))
	}
}
