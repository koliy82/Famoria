package message

type Repository interface {
	Insert(message *Message)
	MessageCount(userID int64, chatID int64) (uint64, error)
}
