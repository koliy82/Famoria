package predicate

import (
	"context"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// IsChatAdmin reports whether the user is an administrator or the creator of the
// chat, according to the Telegram API.
//
// It returns false on any API error, so a failed lookup never grants access.
//
// Command handlers call this directly rather than through a predicate, because
// a rejected caller should be told why instead of getting silence.
func IsChatAdmin(ctx context.Context, bot *telego.Bot, chatID, userID int64) bool {
	member, err := bot.GetChatMember(ctx, &telego.GetChatMemberParams{
		ChatID: tu.ID(chatID),
		UserID: userID,
	})
	if err != nil {
		return false
	}
	switch member.MemberStatus() {
	case telego.MemberStatusCreator, telego.MemberStatusAdministrator:
		return true
	default:
		return false
	}
}
