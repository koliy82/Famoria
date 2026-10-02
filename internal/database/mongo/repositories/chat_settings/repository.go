package chat_settings

import (
	"famoria/internal/pkg/i18n"
)

// Repository provides per-chat bot configuration.
type Repository interface {
	// Get returns the settings for a chat, never nil. Chats without a stored
	// document get the defaults.
	Get(chatID int64) *ChatSettings

	// Lang returns the display language configured for a chat.
	Lang(chatID int64) i18n.Lang

	// IsVideoConverterEnabled reports whether link-to-video conversion is on.
	IsVideoConverterEnabled(chatID int64) bool

	// DeleteOriginalLink reports whether the source link message is removed
	// after a video was sent successfully.
	DeleteOriginalLink(chatID int64) bool

	// IsFeatureEnabled reports whether a toggleable feature is on for a chat.
	IsFeatureEnabled(chatID int64, f Feature) bool

	// SetLang stores the display language for a chat.
	SetLang(chatID int64, lang i18n.Lang)

	// SetVideoConverter turns link-to-video conversion on or off.
	SetVideoConverter(chatID int64, enabled bool)

	// SetVideoKeepOriginal controls whether the source link message is kept
	// after a successful upload.
	SetVideoKeepOriginal(chatID int64, keep bool)

	// SetFeatureEnabled turns a toggleable feature on or off.
	SetFeatureEnabled(chatID int64, f Feature, enabled bool)
}
