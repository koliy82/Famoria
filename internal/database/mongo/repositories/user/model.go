package user

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	OID          primitive.ObjectID `bson:"_id"`
	ID           int64              `bson:"id"`
	FirstName    string             `bson:"first_name"`
	LastName     *string            `bson:"last_name"`
	Username     *string            `bson:"username"`
	LanguageCode string             `bson:"language_code"`
	IsAdmin      bool               `bson:"is_admin"`
}

func (u *User) IsEquals(other *User) bool {
	return u.ID == other.ID &&
		u.FirstName == other.FirstName &&
		u.LastName == other.LastName &&
		u.Username == other.Username &&
		u.LanguageCode == other.LanguageCode
}
