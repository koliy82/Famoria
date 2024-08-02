package html

import (
	"fmt"
	"github.com/mymmrac/telego"
	"go_tg_bot/internal/database/mongo/repositories/user"
	"strings"
)

// Bold applies the bold font style to the string.
// Passed string will not be automatically escaped because it can contain nested markup.
func Bold(s string) string {
	return fmt.Sprintf("<b>%s</b>", s)
}

// Italic applies the italic font style to the string.
// Passed string will not be automatically escaped because it can contain nested markup.
func Italic(s string) string {
	return fmt.Sprintf("<i>%s</i>", s)
}

// Underline applies the underline font style to the string.
// Passed string will not be automatically escaped because it can contain nested markup.
func Underline(s string) string {
	return fmt.Sprintf("<u>%s</u>", s)
}

// Strike applies the strike font style to the string.
// Passed string will not be automatically escaped because it can contain nested markup.
func Strike(s string) string {
	return fmt.Sprintf("<s>%s</s>", s)
}

// Link creates a hyperlink with the specified URL and text.
// Passed strings will be automatically escaped.
func Link(url string, text string) string {
	return fmt.Sprintf("<a href=\"%s\">%s</a>", Escape(url), Escape(text))
}

// Mention creates a mention of the user with the specified ID and text.
// Passed string will be automatically escaped.
func Mention(id int64, text string) string {
	return Link(fmt.Sprintf("tg://user?id=%d", id), text)
}

// MentionURL creates a mention of the user with the specified username and text.
func MentionURL(username string, text string) string {
	return Link("https://t.me/"+username, text)
}

// UserMention builds an inline user mention link with an anchor.
func UserMention(user *telego.User) string {
	if &user.Username != nil {
		return Mention(user.ID, user.Username)
	}
	if &user.LastName != nil {
		return Mention(user.ID, fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	}
	return Mention(user.ID, user.FirstName)
}

// ModelMention creates a mention of the user with the specified ID and text.
func ModelMention(user *user.User) string {
	if &user.Username != nil {
		return Mention(user.ID, *user.Username)
	}
	if &user.LastName != nil {
		return Mention(user.ID, fmt.Sprintf("%s %s", user.FirstName, *user.LastName))
	}
	return Mention(user.ID, user.FirstName)
}

// CodeBlock creates a code block with the specified code.
// Escapes HTML characters inside the block.
func CodeBlock(code string) string {
	return fmt.Sprintf("<pre>%s</pre>", Escape(code))
}

// CodeBlockWithLang creates a code block with the specified code and language.
// Escapes HTML characters inside the block.
func CodeBlockWithLang(code, lang string) string {
	return fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>",
		strings.Replace(Escape(lang), "\"", "&quot;", -1), Escape(code),
	)
}

// CodeInline creates an inline code block with the specified code.
// Escapes HTML characters inside the block.
func CodeInline(s string) string {
	return fmt.Sprintf("<code>%s</code>", Escape(s))
}

// Escape the string to be shown "as is" within the Telegram HTML message style.
func Escape(s string) string {
	res := strings.Replace(s, "&", "&amp;", -1)
	res = strings.Replace(res, "<", "&lt;", -1)
	res = strings.Replace(res, ">", "&gt;", -1)
	return res
}
