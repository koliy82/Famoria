package message

type Repository interface {
	Insert(message *Message)
}
