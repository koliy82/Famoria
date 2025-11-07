package message

import "go.mongodb.org/mongo-driver/bson/primitive"

type Message struct {
	OID     primitive.ObjectID `bson:"_id"`
	ID      int                `bson:"id"`
	ChatID  int64              `bson:"chat_id"`
	UserID  int64              `bson:"user_id"`
	Date    int64              `bson:"date"`
	Text    *string            `bson:"text"`
	Caption *string            `bson:"caption"`
	ReplyID *int               `bson:"reply_id"`
}
