package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go_tg_bot/internal/config"
)

func New(cfg config.Config) *mongo.Client {
	client, err := mongo.Connect(
		context.TODO(), options.Client().ApplyURI(cfg.MongoURI),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	return client
}
