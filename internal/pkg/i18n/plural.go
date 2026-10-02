package i18n

// Pluralized phrases. Call sites use these helpers instead of passing word
// forms themselves, so no caller needs to know how a language inflects.

// Marriages returns the header line for the marriage list.
func Marriages(lang Lang, count int64, global bool) string {
	key := KeyBraksHeaderLocal
	if global {
		key = KeyBraksHeaderGlobal
	}

	switch lang {
	case LangEN:
		return T(lang, key, count, Plural(lang, count, enBrakOne, "", enBrakMany))
	default:
		return T(lang, key, count, Plural(lang, count, ruBrakOne, ruBrakFew, ruBrakMany))
	}
}

// Days returns a day count with the correctly inflected word, e.g. "3 дня".
func Days(lang Lang, n int64) string {
	switch lang {
	case LangEN:
		return Plural(lang, n, enDayOne, "", enDayMany)
	default:
		return Plural(lang, n, ruDayOne, ruDayFew, ruDayMany)
	}
}
