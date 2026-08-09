package chat_settings

type ChatSettings struct {
	ChatID                int64 `bson:"chat_id"`
	VideoConverterEnabled bool  `bson:"video_converter_enabled"`
}
