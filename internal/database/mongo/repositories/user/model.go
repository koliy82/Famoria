package user

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go_tg_bot/internal/utils/html"
)

type User struct {
	_ID          primitive.ObjectID `bson:"_id"`
	ID           int64              `ch:"id"`
	FirstName    string             `ch:"first_name"`
	LastName     *string            `ch:"last_name"`
	Username     *string            `ch:"username"`
	LanguageCode string             `ch:"language_code"`
	IsAdmin      bool               `ch:"is_admin"`
}

func (u *User) IsEquals(other *User) bool {
	return u.ID == other.ID &&
		u.FirstName == other.FirstName &&
		u.LastName == other.LastName &&
		u.Username == other.Username &&
		u.LanguageCode == other.LanguageCode
}

func (u *User) Mention() string {
	if u.Username != nil {
		return html.Mention(u.ID, *u.Username)
	}
	if u.LastName != nil {
		return html.Mention(u.ID, fmt.Sprintf("%s %s", u.FirstName, u.LastName))
	}
	return html.Mention(u.ID, u.FirstName)
}
