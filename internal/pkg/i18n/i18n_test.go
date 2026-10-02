package i18n

import "testing"

func TestDictionariesHaveSameKeys(t *testing.T) {
	for lang, d := range dict {
		if lang == DefaultLang {
			continue
		}
		for key := range dict[DefaultLang] {
			if _, ok := d[key]; !ok {
				t.Errorf("lang %s is missing key %q present in %s", lang, key, DefaultLang)
			}
		}
		for key := range d {
			if _, ok := dict[DefaultLang][key]; !ok {
				t.Errorf("lang %s has key %q that %s lacks", lang, key, DefaultLang)
			}
		}
	}
}

func TestTDoesNotReturnEmpty(t *testing.T) {
	for lang := range dict {
		for key := range dict[lang] {
			if got := T(lang, key); got == "" {
				t.Errorf("lang %s key %q resolved to empty string", lang, key)
			}
		}
	}
}

func TestTUnknownKeyFallsBackToKey(t *testing.T) {
	if got := T(LangEN, "no.such.key"); got != "no.such.key" {
		t.Errorf("expected unknown key to be returned verbatim, got %q", got)
	}
}

func TestParseLang(t *testing.T) {
	cases := map[string]Lang{
		"":          DefaultLang,
		"ru":        LangRU,
		"RU":        LangRU,
		"English":   LangEN,
		"en":        LangEN,
		"  en-US  ": LangEN,
		"de":        DefaultLang,
		"garbage":   DefaultLang,
	}
	for in, want := range cases {
		if got := ParseLang(in); got != want {
			t.Errorf("ParseLang(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestPlural(t *testing.T) {
	ruCases := map[int64]string{
		1: ruBrakOne, 21: ruBrakOne, 101: ruBrakOne,
		2: ruBrakFew, 4: ruBrakFew, 22: ruBrakFew,
		5: ruBrakMany, 11: ruBrakMany, 12: ruBrakMany, 0: ruBrakMany, 100: ruBrakMany,
	}
	for n, want := range ruCases {
		if got := Plural(LangRU, n, ruBrakOne, ruBrakFew, ruBrakMany); got != want {
			t.Errorf("RU plural(%d) = %s, want %s", n, got, want)
		}
	}

	if got := Plural(LangEN, 1, enBrakOne, "", enBrakMany); got != enBrakOne {
		t.Errorf("EN plural(1) = %s, want %s", got, enBrakOne)
	}
	for _, n := range []int64{0, 2, 5, 11, 21} {
		if got := Plural(LangEN, n, enBrakOne, "", enBrakMany); got != enBrakMany {
			t.Errorf("EN plural(%d) = %s, want %s", n, got, enBrakMany)
		}
	}
}

func TestPluralHelpers(t *testing.T) {
	if got := Marriages(LangRU, 3, false); got == "" {
		t.Error("Marriages returned empty string")
	}
	if got := Marriages(LangEN, 3, true); got == "" {
		t.Error("Marriages returned empty string")
	}
	if got := Days(LangRU, 3); got != ruDayFew {
		t.Errorf("Days(ru,3) = %q, want %q", got, ruDayFew)
	}
	if got := Days(LangEN, 3); got != enDayMany {
		t.Errorf("Days(en,3) = %q, want %q", got, enDayMany)
	}
}

func TestLangNameIsValid(t *testing.T) {
	for _, l := range SupportedLangs {
		if !l.IsValid() {
			t.Errorf("supported lang %s reports as invalid", l)
		}
		if LangName(l) == "" {
			t.Errorf("LangName(%s) is empty", l)
		}
	}
}
