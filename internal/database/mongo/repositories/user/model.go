package user

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	_ID          primitive.ObjectID `bson:"_id"`
	ID           int64              `ch:"id"`
	FirstName    string             `ch:"first_name"`
	LastName     string             `ch:"last_name"`
	Username     string             `ch:"username"`
	LanguageCode string             `ch:"language_code"`
	IsAdmin      bool               `ch:"is_admin"`
}
