package i18n

// en is the English dictionary.
var en = map[string]string{
	// Settings menu.
	KeySettingsTitle:   "⚙️ <b>Chat settings</b>",
	KeySettingsDesc:    "Choose a category:",
	KeySettingsAdmin:   "This command is available to chat administrators only.",
	KeySettingsUpdated: "Setting updated",
	KeySettingsBack:    "⬅️ Back",

	KeySettingsLangTitle:   "🌐 <b>Language</b>",
	KeySettingsLangDesc:    "Current chat language: <b>%s</b>\n\nChoose the language the bot replies in for this chat.",
	KeySettingsVideoTitle:  "🎬 <b>Video</b>",
	KeySettingsVideoDesc:   "The bot checks video links (YouTube/Shorts/TikTok/Reels and more), downloads them and sends the video into the chat.",
	KeySettingsBrakTitle:   "💍 <b>Marriages</b>",
	KeySettingsBrakDesc:    "Disabled features stop working in this chat.",
	KeySettingsGachaTitle:  "🎰 <b>Gacha</b>",
	KeySettingsGachaDesc:   "Disabled features stop working in this chat.",
	KeySettingsEarnWarning: "⚠️ Turning earnings off removes the job buttons and the balance from the profile, and the marriage list is then sorted by creation date.",

	KeyCatLang:     "🌐 Language",
	KeyCatVideo:    "🎬 Video",
	KeyCatMarriage: "💍 Marriages",
	KeyCatGacha:    "🎰 Gacha",

	// Video converter.
	KeyVideoConverterOn:  "🎬 Video converter: ON",
	KeyVideoConverterOff: "🎬 Video converter: OFF",
	KeyVideoDeleteOn:     "🗑 Delete source link: ON",
	KeyVideoDeleteOff:    "🗑 Delete source link: OFF",

	// Features.
	KeyFeatureBraks:         "💬 Chat marriages",
	KeyFeatureBraksGlobal:   "🌍 All marriages",
	KeyFeatureGoBrak:        "💒 Getting married",
	KeyFeatureEndBrak:       "💔 Divorce",
	KeyFeatureKid:           "👶 Having a baby",
	KeyFeatureDetdom:        "🏠 Orphanage",
	KeyFeatureKidAnnihilate: "☄️ Annihilate the baby",
	KeyFeatureTree:          "🌱 Family tree",
	KeyFeatureShop:          "🛒 Shop",
	KeyFeatureInventory:     "🎒 Inventory",
	KeyFeatureEarnings:      "💰 Chat earnings",

	// Generic.
	KeyFeatureDisabled: "This feature is disabled by the chat administrators.",

	// Profile.
	KeyProfileTitle:    "Profile",
	KeyProfileMessages: "💬 %d",
	KeyProfileMarriage: "Marriage",
	KeyProfileSubDays:  "💎 Subscription for %d %s",
	KeyProfileNoSub:    "😿 No active subscription",

	// Help and menu.
	KeyHelpIntro:    "The bot is built around creating families between users, so most of its functionality becomes available after you marry another user.\n\nAvailable commands:",
	KeyHelpCommands: "/%s - %s",
	KeyMenuShown:    "Menu is shown ✅",
	KeyMenuClosed:   "Menu closed, you can reopen it with /menu.",

	// Braks.
	KeyBraksHeaderLocal:  "💍 %d %s IN THE GROUP 💍\n",
	KeyBraksHeaderGlobal: "💍 %d %s IN THE CHATS 💍\n",
	KeyBraksEmpty:        "There are no marriages in this chat",
	KeyBraksError:        "An error occurred while fetching the marriage list",
	KeyBraksPageHint:     "Page %d",
	KeyBraksDuration:     "⏳ %s",
	KeyBraksAnd:          " and ",

	// Gacha.
	KeyShopNeedMarriage:      "You need to be married to view the marriage shop.",
	KeyInventoryNeedMarriage: "You need to be married to view the marriage inventory.",
	KeyInventoryChoose:       "Choose an item to view.\n",

	// Reply keyboard.
	KeyBtnProfile:    "👤 Profile",
	KeyBtnDivorce:    "💔 Divorce",
	KeyBtnAnnihilate: "👶 Annihilate",
	KeyBtnOrphanage:  "🏠 Orphanage",
	KeyBtnTree:       "🌱 Family tree",
	KeyBtnClose:      "❌ Close",
	KeyBtnShop:       "🛒 Shop",
	KeyBtnInventory:  "🎒 Inventory",
	KeyBtnSteam:      "🎮 Steam accounts",
	KeyBtnSubscribe:  "💳 Subscription",
}

// Plural word forms used by the English dictionary.
const (
	enBrakOne  = "MARRIAGE"
	enBrakMany = "MARRIAGES"

	enDayOne  = "day"
	enDayMany = "days"
)

// dict maps every supported language to its dictionary.
var dict = map[Lang]map[string]string{
	LangRU: ru,
	LangEN: en,
}

// LangName returns the language name in the language itself, for display in
// the settings menu.
func LangName(l Lang) string {
	switch l {
	case LangEN:
		return "English"
	case LangRU:
		return "Русский"
	default:
		return string(l)
	}
}
