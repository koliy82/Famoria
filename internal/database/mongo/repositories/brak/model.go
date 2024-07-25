package brak

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Brak struct {
	BID            primitive.ObjectID `bson:"_id"`
	FirstUserID    int64              `bson:"first_user_id"`
	SecondUserID   int64              `bson:"second_user_id"`
	CreateDate     time.Time          `bson:"create_date"`
	BabyUserID     *int64             `bson:"baby_user_id"`
	BabyCreateDate *time.Time         `bson:"baby_create_date"`
	Score          int64              `bson:"score"`
}
