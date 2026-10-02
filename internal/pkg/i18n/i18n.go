// Package i18n provides chat-localized strings for the bot.
//
// Translations are keyed lookups: T(lang, key, args...) resolves a key in the
// requested language, falling back to the default language and finally to the
// key itself so a missing translation is visible rather than silently blank.
package i18n

import (
	"fmt"
	"strings"
)

// Lang is a supported chat language.
type Lang string

const (
	LangRU Lang = "ru"
	LangEN Lang = "en"
)

// DefaultLang is used for chats without an explicit language setting and as the
// fallback when a key is missing from the requested language.
const DefaultLang = LangRU

// SupportedLangs lists every language selectable in chat settings, in display
// order.
var SupportedLangs = []Lang{LangRU, LangEN}

// ParseLang normalizes a language code from config or storage. Unknown and empty
// values resolve to DefaultLang.
func ParseLang(s string) Lang {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(LangEN), "english", "eng", "en-us", "en-gb":
		return LangEN
	case string(LangRU), "russian", "rus", "ru-ru", "":
		return LangRU
	default:
		return DefaultLang
	}
}

func (l Lang) String() string { return string(l) }

// IsValid reports whether l is a supported language.
func (l Lang) IsValid() bool {
	for _, s := range SupportedLangs {
		if s == l {
			return true
		}
	}
	return false
}

// lookup resolves a key in the given language, then in the default language.
func lookup(lang Lang, key string) (string, bool) {
	if d, ok := dict[lang]; ok {
		if s, ok := d[key]; ok {
			return s, true
		}
	}
	if lang != DefaultLang {
		if d, ok := dict[DefaultLang]; ok {
			if s, ok := d[key]; ok {
				return s, true
			}
		}
	}
	return "", false
}

// T returns the translation for key in lang. When args are provided the
// translation is used as a fmt.Sprintf format string.
func T(lang Lang, key string, args ...any) string {
	s, ok := lookup(lang, key)
	if !ok {
		return key
	}
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

// Variants returns key translated into every supported language, deduplicated
// and in SupportedLangs order.
//
// Reply-keyboard buttons are dispatched by their literal text, so a chat that
// switched language sends the new label. Handlers must accept every variant of a
// label, not just the one in the default language.
func Variants(key string) []string {
	seen := make(map[string]struct{}, len(SupportedLangs))
	out := make([]string, 0, len(SupportedLangs))
	for _, l := range SupportedLangs {
		s := T(l, key)
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// Plural returns the correct word form for n in lang.
//
// Russian distinguishes one/few/many; English only one/other, so many is used
// as the "other" form and few is ignored.
func Plural(lang Lang, n int64, one, few, many string) string {
	if lang == LangEN {
		if n == 1 {
			return one
		}
		return many
	}

	a := n
	if a < 0 {
		a = -a
	}
	switch {
	case a%10 == 1 && a%100 != 11:
		return one
	case a%10 >= 2 && a%10 <= 4 && (a%100 < 12 || a%100 > 14):
		return few
	default:
		return many
	}
}
