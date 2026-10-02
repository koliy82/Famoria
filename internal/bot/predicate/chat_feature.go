package predicate

import (
	"context"

	"famoria/internal/database/mongo/repositories/chat_settings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// ChatFeatureEnabled gates a handler on a per-chat feature switch. It returns
// false (so the handler never runs) when the chat's administrators turned the
// feature off.
//
// Callback queries and messages without a chat are rejected, because there is
// nothing to authorize them against.
func ChatFeatureEnabled(repo chat_settings.Repository, f chat_settings.Feature) th.Predicate {
	return func(_ context.Context, update telego.Update) bool {
		chatID, ok := updateChatID(update)
		if !ok {
			return false
		}
		return repo.IsFeatureEnabled(chatID, f)
	}
}

// ChatFeaturesEnabled gates a handler on several feature switches at once; all
// of them must be enabled.
func ChatFeaturesEnabled(repo chat_settings.Repository, features ...chat_settings.Feature) th.Predicate {
	return func(_ context.Context, update telego.Update) bool {
		chatID, ok := updateChatID(update)
		if !ok {
			return false
		}
		for _, f := range features {
			if !repo.IsFeatureEnabled(chatID, f) {
				return false
			}
		}
		return true
	}
}

// updateChatID extracts the chat an update belongs to, covering both plain
// messages and inline callback queries.
func updateChatID(update telego.Update) (int64, bool) {
	switch {
	case update.Message != nil:
		return update.Message.Chat.ID, true
	case update.EditedMessage != nil:
		return update.EditedMessage.Chat.ID, true
	case update.CallbackQuery != nil && update.CallbackQuery.Message != nil:
		return update.CallbackQuery.Message.GetChat().ID, true
	default:
		return 0, false
	}
}
