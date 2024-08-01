package message

type Message struct {
	ID      int     `ch:"id"`
	ChatID  int64   `ch:"chat_id"`
	UserID  int64   `ch:"user_id"`
	Date    int64   `ch:"date"`
	Text    *string `ch:"text"`
	Caption *string `ch:"caption"`
	ReplyID *int    `ch:"reply_id"`
}
