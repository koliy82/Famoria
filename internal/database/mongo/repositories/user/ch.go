package user

import (
	"context"
	"errors"
	"github.com/mymmrac/telego"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Ch struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func (c *Ch) ValidateInfo(user *telego.User) error {
	model := &User{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     &user.LastName,
		Username:     &user.Username,
		LanguageCode: user.LanguageCode,
		IsAdmin:      false,
	}
	actual, _ := c.FindByID(model.ID)
	if actual == nil {
		err := c.Insert(model)
		if err != nil {
			return err
		}
		return nil
	}
	if !model.IsEquals(actual) {
		_ = c.Replace(model)
	}
	return nil
}

func (c *Ch) Replace(user *User) error {
	filter := bson.D{{"id", user.ID}}
	_, err := c.coll.ReplaceOne(context.TODO(), filter, user)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func (c *Ch) Insert(user *User) error {
	_, err := c.coll.InsertOne(context.TODO(), user)
	if err != nil {
		c.log.Sugar().Error(err)
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
		c.log.Sugar().Error(err)
	}

	return user, nil
}

func New(client *mongo.Client, log *zap.Logger) *Ch {
	return &Ch{
		coll: client.Database("test").Collection("users"),
		log:  log,
	}
}
