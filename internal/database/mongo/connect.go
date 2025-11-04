package mongo

import (
	"context"
	"famoria/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func New(cfg config.Config) *mongo.Client {
	client, err := mongo.Connect(
		context.TODO(), options.Client().ApplyURI(cfg.MongoURI),
	)
	if err != nil {
		panic(err)
	}
	return client
}
