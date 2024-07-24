package brak

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Ch struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func (c *Ch) FindByUserID(id int64) (*Brak, error) {
	result := &Brak{}
	filter := bson.D{
		{"$or", []interface{}{
			bson.D{{"firstuserid", id}},
			bson.D{{"seconduserid", id}},
		}},
	}
	err := c.coll.FindOne(context.TODO(), filter).Decode(result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		c.log.Sugar().Error(err)
	}
	return result, nil
}

func (c *Ch) Insert(brak *Brak) error {
	_, err := c.coll.InsertOne(context.TODO(), brak)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func (c *Ch) Delete(id primitive.ObjectID) error {
	_, err := c.coll.DeleteOne(context.TODO(), bson.D{{"_id", id}})
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func New(client *mongo.Client, log *zap.Logger) *Ch {
	return &Ch{
		coll: client.Database("test").Collection("braks"),
		log:  log,
	}
}
