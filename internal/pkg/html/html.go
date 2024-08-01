package html

import (
	"fmt"
	"github.com/mymmrac/telego"
)

func Bold(s string) string {
	return fmt.Sprintf("<b>%s</b>", s)
}

func Italic(s string) string {
	return fmt.Sprintf("<i>%s</i>", s)
}

func Underline(s string) string {
	return fmt.Sprintf("<u>%s</u>", s)
}

func Strike(s string) string {
	return fmt.Sprintf("<s>%s</s>", s)
}

func Link(url string, text string) string {
	return fmt.Sprintf("<a href=\"%s\">%s</a>", url, text)
}

func Mention(id int64, text string) string {
	return Link(fmt.Sprintf("tg://user?id=%d", id), text)
}

func UserMention(user *telego.User) string {
	if &user.Username != nil {
		return Mention(user.ID, user.Username)
	}
	if &user.LastName != nil {
		return Mention(user.ID, fmt.Sprintf("%s %s", user.FirstName, user.LastName))
	}
	return Mention(user.ID, user.FirstName)
}

func CodeBlock(code string) string {
	return fmt.Sprintf("<pre>%s</pre>", code)
}

func CodeBlockWithLang(code, lang string) string {
	return fmt.Sprintf("<pre><code class=\"language-%s\">%s</code></pre>", lang, code)
}

func CodeInline(s string) string {
	return fmt.Sprintf("<code>%s</code>", s)
}
