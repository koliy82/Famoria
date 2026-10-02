package chat_settings

import (
	"strings"

	"famoria/internal/pkg/i18n"
)

// Feature identifies a toggleable bot function within a chat.
type Feature string

// Marriage features, shown in the "Браки / Marriages" settings category.
const (
	FeatBraks         Feature = "braks"
	FeatBraksGlobal   Feature = "braksglobal"
	FeatGoBrak        Feature = "gobrak"
	FeatEndBrak       Feature = "endbrak"
	FeatKid           Feature = "kid"
	FeatDetdom        Feature = "detdom"
	FeatKidAnnihilate Feature = "kidannihilate"
	FeatTree          Feature = "tree"
)

// Gacha features, shown in the "Гача / Gacha" settings category.
const (
	FeatShop      Feature = "shop"
	FeatInventory Feature = "inventory"
	FeatEarnings  Feature = "earnings"
)

// marriageFeatures lists the toggleable marriage features in display order.
var marriageFeatures = []Feature{
	FeatBraks,
	FeatBraksGlobal,
	FeatGoBrak,
	FeatEndBrak,
	FeatKid,
	FeatDetdom,
	FeatKidAnnihilate,
	FeatTree,
}

// gachaFeatures lists the toggleable gacha features in display order.
var gachaFeatures = []Feature{
	FeatShop,
	FeatInventory,
	FeatEarnings,
}

// MarriageFeatures returns the toggleable marriage features in display order.
func MarriageFeatures() []Feature { return append([]Feature(nil), marriageFeatures...) }

// GachaFeatures returns the toggleable gacha features in display order.
func GachaFeatures() []Feature { return append([]Feature(nil), gachaFeatures...) }

// i18nKey maps a feature to the translation key of its button label.
func (f Feature) i18nKey() string {
	switch f {
	case FeatBraks:
		return i18n.KeyFeatureBraks
	case FeatBraksGlobal:
		return i18n.KeyFeatureBraksGlobal
	case FeatGoBrak:
		return i18n.KeyFeatureGoBrak
	case FeatEndBrak:
		return i18n.KeyFeatureEndBrak
	case FeatKid:
		return i18n.KeyFeatureKid
	case FeatDetdom:
		return i18n.KeyFeatureDetdom
	case FeatKidAnnihilate:
		return i18n.KeyFeatureKidAnnihilate
	case FeatTree:
		return i18n.KeyFeatureTree
	case FeatShop:
		return i18n.KeyFeatureShop
	case FeatInventory:
		return i18n.KeyFeatureInventory
	case FeatEarnings:
		return i18n.KeyFeatureEarnings
	default:
		return string(f)
	}
}

// Label returns the localized display name of the feature.
func (f Feature) Label(lang i18n.Lang) string {
	return i18n.T(lang, f.i18nKey())
}

// normalizeFeature maps a raw string to a known Feature. Unknown values resolve
// to the empty string so they cannot silently disable a real feature.
func normalizeFeature(s string) Feature {
	f := Feature(strings.ToLower(strings.TrimSpace(s)))
	for _, known := range append(append([]Feature{}, marriageFeatures...), gachaFeatures...) {
		if known == f {
			return f
		}
	}
	return ""
}

// ChatSettings stores per-chat bot configuration.
//
// DisabledFeatures stores only the features that are OFF. Absent keys mean
// enabled, which keeps documents written before this field existed behaving as
// they always have — no migration required.
type ChatSettings struct {
	ChatID int64 `bson:"chat_id"`

	// Lang is the chat's display language, stored as the i18n language code.
	Lang string `bson:"lang"`

	// VideoConverterEnabled turns the link-to-video converter on for the chat.
	VideoConverterEnabled bool `bson:"video_converter_enabled"`

	// VideoKeepOriginal keeps the source link message after the video was sent
	// successfully.
	//
	// This is stored inverted (keep, not delete) on purpose: documents written
	// before this field existed decode it as false, which means "delete the
	// original" — the behaviour the converter has always had. Storing it the
	// other way round would silently stop deleting links in every chat that
	// already had the converter enabled.
	VideoKeepOriginal bool `bson:"video_keep_original"`

	// DisabledFeatures holds the features switched off in this chat.
	DisabledFeatures map[string]bool `bson:"disabled_features"`
}

// newDefault builds the default settings for a chat that has no stored document.
func newDefault(chatID int64) *ChatSettings {
	return &ChatSettings{
		ChatID:                chatID,
		Lang:                  string(i18n.DefaultLang),
		VideoConverterEnabled: false,
		VideoKeepOriginal:     false,
		DisabledFeatures:      make(map[string]bool),
	}
}

// Language returns the chat's configured language, falling back to the default.
func (s *ChatSettings) Language() i18n.Lang {
	l := i18n.ParseLang(s.Lang)
	if !l.IsValid() {
		return i18n.DefaultLang
	}
	return l
}

// T resolves a translation key in the chat's language.
func (s *ChatSettings) T(key string, args ...any) string {
	return i18n.T(s.Language(), key, args...)
}

// IsFeatureEnabled reports whether a feature is switched on for this chat.
// Unknown features are treated as enabled so an unrecognised value never locks
// functionality away by accident.
func (s *ChatSettings) IsFeatureEnabled(f Feature) bool {
	if normalizeFeature(string(f)) == "" {
		return true
	}
	return !s.DisabledFeatures[string(f)]
}

// DeleteOriginalLink reports whether the source link should be deleted after a
// successful video upload.
func (s *ChatSettings) DeleteOriginalLink() bool {
	return !s.VideoKeepOriginal
}

// clone returns a deep copy so callers cannot mutate cached state.
func (s *ChatSettings) clone() *ChatSettings {
	cp := *s
	cp.DisabledFeatures = make(map[string]bool, len(s.DisabledFeatures))
	for k, v := range s.DisabledFeatures {
		cp.DisabledFeatures[k] = v
	}
	return &cp
}
