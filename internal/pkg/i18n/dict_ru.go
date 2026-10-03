package i18n

// ru is the Russian dictionary and the source of truth for every key.
var ru = map[string]string{
	// Settings menu.
	KeySettingsTitle:   "⚙️ <b>Настройки чата</b>",
	KeySettingsDesc:    "Выберите категорию:",
	KeySettingsAdmin:   "Команда доступна только администраторам чата.",
	KeySettingsUpdated: "Настройка обновлена",
	KeySettingsBack:    "⬅️ Назад",

	KeySettingsLangTitle:   "🌐 <b>Язык</b>",
	KeySettingsLangDesc:    "Текущий язык чата: <b>%s</b>\n\nВыберите язык, на котором бот будет отвечать в этом чате.",
	KeySettingsVideoTitle:  "🎬 <b>Видео</b>",
	KeySettingsVideoDesc:   "Бот проверяет ссылки на видео (YouTube/Shorts/TikTok/Reels и др.), скачивает и отправляет видео в чат.",
	KeySettingsBrakTitle:   "💍 <b>Браки</b>",
	KeySettingsBrakDesc:    "Отключённые функции перестанут работать в этом чате.",
	KeySettingsGachaTitle:  "🎰 <b>Гача</b>",
	KeySettingsGachaDesc:   "Отключённые функции перестанут работать в этом чате.",
	KeySettingsEarnWarning: "⚠️ При отключении заработка из профиля пропадут кнопки работы и баланс, а список браков будет сортироваться по дате создания.",

	KeyCatLang:     "🌐 Язык",
	KeyCatVideo:    "🎬 Видео",
	KeyCatMarriage: "💍 Браки",
	KeyCatGacha:    "🎰 Гача",

	// Video converter.
	KeyVideoConverterOn:  "🎬 Видео-конвертер: ВКЛ",
	KeyVideoConverterOff: "🎬 Видео-конвертер: ВЫКЛ",
	KeyVideoDeleteOn:     "🗑 Удаление ссылки: ВКЛ",
	KeyVideoDeleteOff:    "🗑 Удаление ссылки: ВЫКЛ",

	// Features.
	KeyFeatureBraks:         "💬 Браки чата",
	KeyFeatureBraksGlobal:   "🌍 Браки всех чатов",
	KeyFeatureGoBrak:        "💒 Заключение брака",
	KeyFeatureEndBrak:       "💔 Развод",
	KeyFeatureKid:           "👶 Рождение ребёнка",
	KeyFeatureDetdom:        "🏠 Детдом",
	KeyFeatureKidAnnihilate: "☄️ Аннигиляция ребёнка",
	KeyFeatureTree:          "🌱 Семейное древо",
	KeyFeatureShop:          "🛒 Магазин",
	KeyFeatureInventory:     "🎒 Инвентарь",
	KeyFeatureEarnings:      "💰 Заработок в чате",

	// Generic.
	KeyFeatureDisabled: "Эта функция отключена администраторами чата.",

	// Profile.
	KeyProfileTitle:    "Профиль",
	KeyProfileMessages: "💬 %d",
	KeyProfileMarriage: "Брак",
	KeyProfileSubDays:  "💎 Подписка на %d %s",
	KeyProfileNoSub:    "😿 Нет активной подписки",

	// Help and menu.
	KeyHelpIntro:    "Основная концепция бота заключается в создании семей между пользователями, поэтому основной функционал бота становится доступен после вступления в брак с другим пользователем.\n\nДоступные команды:",
	KeyHelpCommands: "/%s - %s",
	KeyMenuShown:    "Меню показано ✅",
	KeyMenuClosed:   "Меню закрыто, повторно открыть его можно написав /menu.",

	// Braks.
	KeyBraksHeaderLocal:  "💍 %d %s В ГРУППЕ 💍\n",
	KeyBraksHeaderGlobal: "💍 %d %s В ЧАТАХ 💍\n",
	KeyBraksEmpty:        "В этом чате нет браков",
	KeyBraksError:        "Произошла ошибка при получении списка браков",
	KeyBraksPageHint:     "Страница №%d",
	KeyBraksDuration:     "⏳ %s",
	KeyBraksAnd:          " и ",

	// Gacha.
	KeyShopNeedMarriage:      "Для просмотра магазина брака вам нужно быть в браке.",
	KeyInventoryNeedMarriage: "Для просмотра инвентаря брака вам нужно быть в браке.",
	KeyInventoryChoose:       "Выберите предмет для просмотра.\n",

	// Link converter output.
	KeyVideoCaption:        "<b>%s</b>\n\n%s · converted by Famoria",
	KeyVideoCaptionNoTitle: "%s · converted by Famoria",
	KeyVideoOriginal:       "Источник",

	// Reply keyboard.
	KeyBtnProfile:    "👤 Профиль",
	KeyBtnDivorce:    "💔 Развод",
	KeyBtnAnnihilate: "👶 Аннигиляция",
	KeyBtnOrphanage:  "🏠 Детдом",
	KeyBtnTree:       "🌱 Семейное древо",
	KeyBtnClose:      "❌ Закрыть",
	KeyBtnShop:       "🛒 Магазин",
	KeyBtnInventory:  "🎒 Инвентарь",
	KeyBtnSteam:      "🎮 Steam аккаунты",
	KeyBtnSubscribe:  "💳 Подписка",
}

// Plural word forms used by the Russian dictionary.
const (
	ruBrakOne  = "БРАК"
	ruBrakFew  = "БРАКА"
	ruBrakMany = "БРАКОВ"

	ruDayOne  = "день"
	ruDayFew  = "дня"
	ruDayMany = "дней"
)
