package predicate

import (
	admin2 "famoria/internal/database/mongo/repositories/admin"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func AdminCommand(repo admin2.Repository, level int) th.Predicate {
	return func(update telego.Update) bool {
		from := update.Message.From
		if from == nil {
			return false
		}
		admin := repo.Get(from.ID)
		if admin == nil {
			return false
		}
		return admin.UserID == from.ID && admin.PermissionLevel >= level
	}
}
