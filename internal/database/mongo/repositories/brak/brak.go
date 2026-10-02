package brak

import (
	"famoria/internal/bot/idle/item"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	FindByUserID(id int64, m *item.Manager) (*Brak, error)
	FindByKidID(id int64) (*Brak, error)
	Insert(brak *Brak) error
	Delete(id primitive.ObjectID) error
	Update(filter interface{}, update interface{}) error
	FindBraksByPage(page int64, limit int64, filter interface{}, sort BrakSort) ([]*UsersBrak, int64, error)
	Count(filter interface{}) (int64, error)
	FindAllMining() ([]*Brak, error)
}

// BrakSort selects the ordering used by FindBraksByPage.
type BrakSort int

const (
	// BrakSortByScore orders marriages by their balance, highest first. Used in
	// chats where earnings are enabled.
	BrakSortByScore BrakSort = iota
	// BrakSortByCreateDate orders marriages by how long they have existed,
	// oldest first. Used in chats where earnings are disabled and a score
	// ranking would be meaningless.
	BrakSortByCreateDate
)

// bsonSort maps the sort to the MongoDB $sort document.
func (s BrakSort) bsonSort() interface{} {
	if s == BrakSortByCreateDate {
		return bson.M{"create_date": 1}
	}
	return bson.M{"score": -1}
}
