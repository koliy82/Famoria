package predicate

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func AdminCommand() th.Predicate {
	return func(update telego.Update) bool {
		from := update.Message.From
		if from == nil {
			return false
		}
		return from.ID == 725757421
	}
}
