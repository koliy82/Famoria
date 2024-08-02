package user

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"strings"
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
