package chat_settings

type Repository interface {
	Get(chatID int64) *ChatSettings
	IsVideoConverterEnabled(chatID int64) bool
	SetVideoConverter(chatID int64, enabled bool)
}
