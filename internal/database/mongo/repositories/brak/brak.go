package brak

import "go.mongodb.org/mongo-driver/bson/primitive"

type Repository interface {
	FindByUserID(id int64) (*Brak, error)
	Insert(brak *Brak) error
	Delete(id primitive.ObjectID) error
	UpdateScore(brakID primitive.ObjectID, score int) error
}
