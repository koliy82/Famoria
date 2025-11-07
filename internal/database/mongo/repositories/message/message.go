package message

type Repository interface {
	Insert(message *Message) error
	MessageCount(userID int64, chatID int64) (int64, error)
}
