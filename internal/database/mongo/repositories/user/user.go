package user

type Repository interface {
	FindByID(id int64) (*User, error)
	Insert(user *User) error
}
