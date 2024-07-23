package user

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Ch struct {
	coll *mongo.Collection
}

func (c *Ch) Insert(user *User) error {
	_, err := c.coll.InsertOne(context.TODO(), user)
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *Ch) FindByID(id int64) (*User, error) {
	user := &User{}
	filter := bson.D{{"id", id}}
	//var result User
	err := c.coll.FindOne(context.TODO(), filter).Decode(user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		panic(err)
	}

	return user, nil
}

func New(client *mongo.Client) *Ch {
	return &Ch{
		coll: client.Database("test").Collection("users"),
	}
}
