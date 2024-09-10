package predicate

import (
	admin2 "famoria/internal/database/mongo/repositories/admin"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"strings"
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

func CommandOnlyBotName(command string) th.Predicate {
	return func(update telego.Update) bool {
		if update.Message == nil || !strings.HasPrefix(update.Message.Text, "/") {
			return false
		}
		parts := strings.SplitN(strings.TrimPrefix(update.Message.Text, "/"), " ", 2)
		cmdParts := strings.SplitN(parts[0], "@", 2)
		// Check command
		if !strings.EqualFold(cmdParts[0], command) {
			return false
		}
		// Check bot name if exist
		if len(cmdParts) > 1 && !strings.EqualFold(cmdParts[1], "testABOBA1Bot") {
			return false
		}
		return true
	}
}
