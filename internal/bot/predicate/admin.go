package predicate

import (
	"github.com/koliy82/telego"
	th "github.com/koliy82/telego/telegohandler"
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
