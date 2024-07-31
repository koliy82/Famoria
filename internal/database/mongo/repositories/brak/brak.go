package brak

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	FindByUserID(id int64) (*Brak, error)
	FindByKidID(id int64) (*Brak, error)
	Insert(brak *Brak) error
	Delete(id primitive.ObjectID) error
	Update(filter interface{}, update interface{}) error
	FindBraksByPage(page int64, limit int64, filter interface{}) ([]*UsersBrak, int64, error)
	Count(filter interface{}) (int64, error)
}
