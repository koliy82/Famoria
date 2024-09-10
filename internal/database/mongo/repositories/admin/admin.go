package admin

type Repository interface {
	ActualData()
	Get(userId int64) *Admin
	Add(admin *Admin)
	Remove(userId int64)
}
