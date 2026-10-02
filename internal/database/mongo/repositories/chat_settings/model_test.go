package chat_settings

import (
	"testing"

	"famoria/internal/pkg/i18n"
)

func TestNewDefaults(t *testing.T) {
	s := newDefault(42)

	if s.ChatID != 42 {
		t.Errorf("ChatID = %d, want 42", s.ChatID)
	}
	if s.Language() != i18n.DefaultLang {
		t.Errorf("Language() = %s, want %s", s.Language(), i18n.DefaultLang)
	}
	if s.VideoConverterEnabled {
		t.Error("VideoConverterEnabled should default to false")
	}
	// A document written before VideoKeepOriginal existed decodes it as false,
	// which must still mean "delete the original link" — the long-standing
	// behaviour of the converter.
	if !s.DeleteOriginalLink() {
		t.Error("DeleteOriginalLink() should default to true")
	}
	if s.DisabledFeatures == nil {
		t.Error("DisabledFeatures should be non-nil so it is writable")
	}
}

func TestLanguageFallsBackOnInvalidStoredValue(t *testing.T) {
	cases := map[string]i18n.Lang{
		"":         i18n.DefaultLang,
		"ru":       i18n.LangRU,
		"en":       i18n.LangEN,
		"klingon":  i18n.DefaultLang,
		"\t EN \n": i18n.LangEN,
	}
	for stored, want := range cases {
		s := newDefault(1)
		s.Lang = stored
		if got := s.Language(); got != want {
			t.Errorf("Lang=%q -> Language() = %s, want %s", stored, got, want)
		}
	}
}

func TestIsFeatureEnabled(t *testing.T) {
	s := newDefault(1)

	for _, f := range append(MarriageFeatures(), GachaFeatures()...) {
		if !s.IsFeatureEnabled(f) {
			t.Errorf("feature %s should be enabled by default", f)
		}
	}

	s.DisabledFeatures[string(FeatShop)] = true
	if s.IsFeatureEnabled(FeatShop) {
		t.Error("FeatShop should be disabled")
	}
	if !s.IsFeatureEnabled(FeatInventory) {
		t.Error("FeatInventory should stay enabled")
	}
}

func TestIsFeatureEnabledIgnoresUnknownFeature(t *testing.T) {
	s := newDefault(1)
	// An unrecognised feature must never be reported as disabled, otherwise a
	// typo in the DB would silently switch off unrelated functionality.
	if !s.IsFeatureEnabled(Feature("nonsense")) {
		t.Error("unknown feature should be treated as enabled")
	}

	s.DisabledFeatures["nonsense"] = true
	if !s.IsFeatureEnabled(Feature("nonsense")) {
		t.Error("unknown feature should be treated as enabled even when stored as disabled")
	}
}

func TestCloneIsDeep(t *testing.T) {
	s := newDefault(7)
	s.DisabledFeatures[string(FeatTree)] = true
	s.Lang = "en"

	cp := s.clone()
	cp.DisabledFeatures[string(FeatShop)] = true
	cp.Lang = "ru"
	cp.VideoConverterEnabled = true

	if s.DisabledFeatures[string(FeatShop)] {
		t.Error("mutating the clone must not affect the original map")
	}
	if s.Lang != "en" {
		t.Errorf("original Lang changed to %q", s.Lang)
	}
	if s.VideoConverterEnabled {
		t.Error("original VideoConverterEnabled changed")
	}
}

func TestNormalizeFeature(t *testing.T) {
	cases := map[string]Feature{
		"shop":          FeatShop,
		"SHOP":          FeatShop,
		"  tree  ":      FeatTree,
		"kidannihilate": FeatKidAnnihilate,
		"":              "",
		"bogus":         "",
	}
	for in, want := range cases {
		if got := normalizeFeature(in); got != want {
			t.Errorf("normalizeFeature(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFeatureLabelsAreTranslated(t *testing.T) {
	for _, f := range append(MarriageFeatures(), GachaFeatures()...) {
		for _, lang := range i18n.SupportedLangs {
			label := f.Label(lang)
			if label == "" {
				t.Errorf("feature %s has empty label in %s", f, lang)
			}
			// A missing translation falls back to the raw key, which is the
			// i18n key itself here — catch that so labels never leak keys.
			if label == f.i18nKey() {
				t.Errorf("feature %s label in %s fell back to its key %q", f, lang, label)
			}
		}
	}
}

func TestFeatureSlicesAreCopies(t *testing.T) {
	a := MarriageFeatures()
	if len(a) == 0 {
		t.Fatal("MarriageFeatures should not be empty")
	}
	a[0] = FeatShop
	if MarriageFeatures()[0] == FeatShop {
		t.Error("MarriageFeatures must return a copy, not the shared slice")
	}

	b := GachaFeatures()
	if len(b) == 0 {
		t.Fatal("GachaFeatures should not be empty")
	}
	b[0] = FeatTree
	if GachaFeatures()[0] == FeatTree {
		t.Error("GachaFeatures must return a copy, not the shared slice")
	}
}

// TestFeatureConstantsAreUnique guards against two features sharing a storage
// key, which would make toggling one silently toggle the other.
func TestFeatureConstantsAreUnique(t *testing.T) {
	all := append(MarriageFeatures(), GachaFeatures()...)
	seen := make(map[Feature]bool, len(all))
	for _, f := range all {
		if seen[f] {
			t.Errorf("duplicate feature %q", f)
		}
		seen[f] = true
	}
	if len(all) != len(seen) {
		t.Errorf("expected %d unique features, got %d", len(all), len(seen))
	}
}
