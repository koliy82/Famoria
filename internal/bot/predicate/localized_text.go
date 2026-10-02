package predicate

import (
	"context"

	"famoria/internal/pkg/i18n"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// TextEqualKey matches a reply-keyboard button by its translation key.
//
// Reply-keyboard buttons are dispatched by their literal text, and that text is
// now rendered in the chat's language. A handler must therefore accept every
// language variant of a label, otherwise switching a chat to English would
// silently break its buttons.
func TextEqualKey(key string) th.Predicate {
	variants := i18n.Variants(key)
	return func(_ context.Context, update telego.Update) bool {
		msg := update.Message
		if msg == nil {
			return false
		}
		for _, v := range variants {
			if msg.Text == v {
				return true
			}
		}
		return false
	}
}

// TextEqualKeys matches any of the given reply-keyboard labels.
func TextEqualKeys(keys ...string) th.Predicate {
	variants := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		variants = append(variants, i18n.Variants(k)...)
	}
	return func(_ context.Context, update telego.Update) bool {
		msg := update.Message
		if msg == nil {
			return false
		}
		for _, v := range variants {
			if msg.Text == v {
				return true
			}
		}
		return false
	}
}
