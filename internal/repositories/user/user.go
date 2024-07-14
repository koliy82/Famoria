package user

import "context"

type Repository interface {
	FindByID(ctx context.Context, id int64) (*User, error)
}
