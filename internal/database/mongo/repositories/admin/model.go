package admin

type Admin struct {
	UserID          int64 `bson:"user_id"`
	PermissionLevel int   `bson:"permission_level"`
}
