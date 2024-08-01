package user

import (
	"github.com/mymmrac/telego"
)

type Repository interface {
	FindByID(id int64) (*User, error)
	Insert(user *User) error
	ValidateInfo(user *telego.User) error
	Replace(user *User) error
}
