package user

import (
	"github.com/mymmrac/telego"
)

type Repository interface {
	FindByID(id int64) (*User, error)
	FindOrUpdate(user *telego.User) (*User, error)
	Insert(user *User) error
	Replace(user *User) error
	Update(filter interface{}, update interface{}) error
}
