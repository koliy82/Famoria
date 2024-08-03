package markdown

import (
	"famoria/internal/database/mongo/repositories/user"
	"fmt"
	"github.com/mymmrac/telego"
	"strings"
)

func Bold(s string) string {
	return fmt.Sprintf("*%s*", s)
}

func Italic(s string) string {
	if len(s) > 1 && s[0:2] == "__" && s[len(s)-2:] == "__" {
		return fmt.Sprintf("_%s\r__", s[:len(s)-1])
	}
	return fmt.Sprintf("_%s_", s)
}

func Underline(s string) string {
	if len(s) > 1 && s[0:1] == "_" && s[len(s)-1:] == "_" {
		return fmt.Sprintf("__%s\r__", s)
	}
	return fmt.Sprintf("__%s__", s)
}

func Strike(s string) string {
	return fmt.Sprintf("~%s~", s)
}

func Link(url string, text string) string {
	return fmt.Sprintf("[%s](%s)", text, EscapeLinkURL(url))
}

func Mention(id int64, text string) string {
	return Link(fmt.Sprintf("tg://user?id=%d", id), text)
}

func MentionURL(username string, text string) string {
	return Link("https://t.me/"+username, text)
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

func ModelMention(user *user.User) string {
	if &user.Username != nil {
		return Mention(user.ID, "@"+*user.Username)
	}
	if &user.LastName != nil {
		return Mention(user.ID, fmt.Sprintf("%s %s", user.FirstName, *user.LastName))
	}
	return Mention(user.ID, user.FirstName)
}

func CodeBlock(code string) string {
	return fmt.Sprintf("```\n%s\n```", EscapeCode(code))
}

func CodeBlockWithLang(code, lang string) string {
	return fmt.Sprintf("```%s\n%s\n```", Escape(lang), EscapeCode(code))
}

func CodeInline(s string) string {
	return fmt.Sprintf("`%s`", EscapeCode(s))
}

func Escape(s string) string {
	res := strings.Replace(s, "_", "\\_", -1)
	res = strings.Replace(res, "*", "\\*", -1)
	res = strings.Replace(res, "[", "\\[", -1)
	res = strings.Replace(res, "]", "\\]", -1)
	res = strings.Replace(res, "(", "\\(", -1)
	res = strings.Replace(res, ")", "\\)", -1)
	res = strings.Replace(res, "~", "\\~", -1)
	res = strings.Replace(res, "`", "\\`", -1)
	res = strings.Replace(res, ">", "\\>", -1)
	res = strings.Replace(res, "#", "\\#", -1)
	res = strings.Replace(res, "+", "\\+", -1)
	res = strings.Replace(res, "-", "\\-", -1)
	res = strings.Replace(res, "=", "\\=", -1)
	res = strings.Replace(res, "|", "\\|", -1)
	res = strings.Replace(res, "{", "\\{", -1)
	res = strings.Replace(res, "}", "\\}", -1)
	res = strings.Replace(res, ".", "\\.", -1)
	res = strings.Replace(res, "!", "\\!", -1)
	return res
}

func EscapeLinkURL(s string) string {
	return strings.Replace(s, "`", "\\`", -1)
}

func EscapeCode(s string) string {
	return strings.Replace(s, "\\", "\\\\", -1)
}
