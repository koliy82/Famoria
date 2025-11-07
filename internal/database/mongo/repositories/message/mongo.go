package message

import (
	"context"
	"famoria/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Mongo struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func (c *Mongo) Insert(ch *Message) error {
	ch.OID = primitive.NewObjectID()
	_, err := c.coll.InsertOne(context.TODO(), ch)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func (c *Mongo) MessageCount(userID int64, chatID int64) (count int64, err error) {
	count, err = c.coll.CountDocuments(context.TODO(), bson.M{"user_id": userID, "chat_id": chatID})
	if err != nil {
		c.log.Error("Error getting message count", zap.Error(err))
		return 0, err
	}
	return count, nil
}

func New(client *mongo.Client, log *zap.Logger, cfg config.Config) *Mongo {
	coll := client.Database(cfg.MongoDatabase).Collection("messages")
	return &Mongo{
		coll: coll,
		log:  log,
	}
}
