package checkout

import (
	"context"
	"famoria/internal/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Mongo struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func (c *Mongo) Insert(ch *Checkout) error {
	ch.OID = primitive.NewObjectID()
	_, err := c.coll.InsertOne(context.TODO(), ch)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func New(client *mongo.Client, log *zap.Logger, cfg config.Config) *Mongo {
	coll := client.Database(cfg.MongoDatabase).Collection("checkouts")
	return &Mongo{
		coll: coll,
		log:  log,
	}
}
