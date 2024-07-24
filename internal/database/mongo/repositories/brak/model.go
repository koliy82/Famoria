package brak

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Brak struct {
	BID            primitive.ObjectID `bson:"_id"`
	FirstUserID    int64              `ch:"first_user_id"`
	SecondUserID   int64              `ch:"second_user_id"`
	CreateDate     time.Time          `ch:"create_date"`
	BabyUserID     *int64             `ch:"baby_user_id"`
	BabyCreateDate *time.Time         `ch:"baby_create_date"`
	Score          int64              `ch:"score"`
}
