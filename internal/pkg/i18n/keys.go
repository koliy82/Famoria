package i18n

// Translation keys, grouped by screen. Using constants keeps call sites
// refactor-safe and makes it easy to spot a key missing from a dictionary.

// Settings menu.
const (
	KeySettingsTitle   = "settings.title"
	KeySettingsDesc    = "settings.desc"
	KeySettingsAdmin   = "settings.admin_only"
	KeySettingsUpdated = "settings.updated"
	KeySettingsBack    = "settings.back"

	KeySettingsLangTitle   = "settings.lang.title"
	KeySettingsLangDesc    = "settings.lang.desc"
	KeySettingsVideoTitle  = "settings.video.title"
	KeySettingsVideoDesc   = "settings.video.desc"
	KeySettingsBrakTitle   = "settings.brak.title"
	KeySettingsBrakDesc    = "settings.brak.desc"
	KeySettingsGachaTitle  = "settings.gacha.title"
	KeySettingsGachaDesc   = "settings.gacha.desc"
	KeySettingsEarnWarning = "settings.earn.warning"

	KeyCatLang     = "settings.cat.lang"
	KeyCatVideo    = "settings.cat.video"
	KeyCatMarriage = "settings.cat.marriage"
	KeyCatGacha    = "settings.cat.gacha"
)

// Video converter settings.
const (
	KeyVideoConverterOn  = "settings.video.converter.on"
	KeyVideoConverterOff = "settings.video.converter.off"
	KeyVideoDeleteOn     = "settings.video.delete.on"
	KeyVideoDeleteOff    = "settings.video.delete.off"
)

// Toggleable features (bracks, gacha, earnings).
const (
	KeyFeatureBraks         = "feature.braks"
	KeyFeatureBraksGlobal   = "feature.braksglobal"
	KeyFeatureGoBrak        = "feature.gobrak"
	KeyFeatureEndBrak       = "feature.endbrak"
	KeyFeatureKid           = "feature.kid"
	KeyFeatureDetdom        = "feature.detdom"
	KeyFeatureKidAnnihilate = "feature.kidannihilate"
	KeyFeatureTree          = "feature.tree"
	KeyFeatureShop          = "feature.shop"
	KeyFeatureInventory     = "feature.inventory"
	KeyFeatureEarnings      = "feature.earnings"
)

// Generic messages.
const (
	KeyFeatureDisabled = "feature.disabled"
)

// Profile.
const (
	KeyProfileTitle    = "profile.title"
	KeyProfileMessages = "profile.messages"
	KeyProfileMarriage = "profile.marriage"
	KeyProfileSubDays  = "profile.sub.days"
	KeyProfileNoSub    = "profile.sub.none"
)

// Help and menu.
const (
	KeyHelpIntro    = "help.intro"
	KeyHelpCommands = "help.commands"
	KeyMenuShown    = "menu.shown"
	KeyMenuClosed   = "menu.closed"
)

// Marriage list (braks).
const (
	KeyBraksHeaderLocal  = "braks.header.local"
	KeyBraksHeaderGlobal = "braks.header.global"
	KeyBraksEmpty        = "braks.empty"
	KeyBraksError        = "braks.error"
	KeyBraksPageHint     = "braks.page_hint"
	KeyBraksDuration     = "braks.duration"
	KeyBraksAnd          = "braks.and"
)

// Gacha (shop / inventory).
const (
	KeyShopNeedMarriage      = "shop.need_marriage"
	KeyInventoryNeedMarriage = "inventory.need_marriage"
	KeyInventoryChoose       = "inventory.choose"
)

// Link converter output.
const (
	// KeyVideoCaption is the caption under converted media: the source title,
	// then the attribution with a link back to where it was posted. The first
	// placeholder is the title, which callers must HTML-escape.
	KeyVideoCaption = "video.caption"
	// KeyVideoCaptionNoTitle is used when the source exposes no title.
	KeyVideoCaptionNoTitle = "video.caption_no_title"
	// KeyVideoOriginal is the clickable label of the source link.
	KeyVideoOriginal = "video.original"
)

// Reply-keyboard button labels. Handlers match every supported language, so
// these labels may change per language without breaking dispatch.
const (
	KeyBtnProfile    = "btn.profile"
	KeyBtnDivorce    = "btn.divorce"
	KeyBtnAnnihilate = "btn.annihilate"
	KeyBtnOrphanage  = "btn.orphanage"
	KeyBtnTree       = "btn.tree"
	KeyBtnClose      = "btn.close"
	KeyBtnShop       = "btn.shop"
	KeyBtnInventory  = "btn.inventory"
	KeyBtnSteam      = "btn.steam"
	KeyBtnSubscribe  = "btn.subscribe"
)
