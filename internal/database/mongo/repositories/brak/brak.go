package brak

type Repository interface {
	FindByUserID(id int64) (*Brak, error)
}
