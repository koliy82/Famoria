package user

import (
	"famoria/internal/pkg/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"strings"
	"time"
)

type User struct {
	OID          primitive.ObjectID `bson:"_id"`
	ID           int64              `bson:"id"`
	FirstName    string             `bson:"first_name"`
	LastName     *string            `bson:"last_name"`
	Username     *string            `bson:"username"`
	LanguageCode string             `bson:"language_code"`
	Score        common.Score       `bson:"score"`
	SubscribeEnd *time.Time         `bson:"subscribe_end"`
}

func (u *User) IsEquals(other *User) bool {
	return strings.EqualFold(u.FirstName, other.FirstName) &&
		reflect.DeepEqual(u.LastName, other.LastName) &&
		reflect.DeepEqual(u.Username, other.Username) &&
		u.LanguageCode == other.LanguageCode
}

func (u *User) UsernameOrFull() string {
	if u.Username != nil {
		return *u.Username
	}
	if u.LastName != nil {
		return u.FirstName + " " + *u.LastName
	}
	return u.FirstName
}

func (u *User) IsSub() bool {
	return u.SubscribeEnd != nil && time.Now().Before(*u.SubscribeEnd)
}

func (u *User) SubDaysCount() int {
	if u.SubscribeEnd == nil {
		return 0
	}
	return int(u.SubscribeEnd.Sub(time.Now()).Hours() / 24)
}
