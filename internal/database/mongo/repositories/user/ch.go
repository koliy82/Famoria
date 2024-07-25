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
	actual, err := c.FindByID(user.ID)
	model := &User{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     &user.LastName,
		Username:     &user.Username,
		LanguageCode: user.LanguageCode,
		IsAdmin:      false,
		MessageCount: 1,
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		err := c.Insert(model)
		if err != nil {
			c.log.Sugar().Error(err)
			return err
		}
		return nil
	}

	filter := bson.M{"id": user.ID}
	if !model.IsEquals(actual) {
		_, err := c.coll.UpdateOne(context.TODO(), filter,
			bson.M{
				"$set": bson.M{
					"first_name":    user.FirstName,
					"last_name":     user.LastName,
					"username":      user.Username,
					"language_code": user.LanguageCode,
				},
				"$inc": bson.M{
					"message_count": 1,
				},
			},
		)
		if err != nil {
			c.log.Sugar().Error(err)
			return err
		}
	} else {
		_, err := c.coll.UpdateOne(
			context.TODO(), filter,
			bson.M{"$inc": bson.M{"message_count": 1}},
		)
		if err != nil {
			c.log.Sugar().Error(err)
			return err
		}
	}
	return nil
}

func (c *Ch) Replace(user *User) error {
	filter := bson.M{"id": user.ID}
	_, err := c.coll.ReplaceOne(context.TODO(), filter, user)
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}
	return nil
}

func (c *Ch) Insert(user *User) error {
	_, err := c.coll.InsertOne(context.TODO(), user)
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}
	return nil
}

func (c *Ch) FindByID(id int64) (*User, error) {
	user := &User{}
	filter := bson.D{{"id", id}}
	//var result User
	err := c.coll.FindOne(context.TODO(), filter).Decode(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func New(client *mongo.Client, log *zap.Logger) *Ch {
	return &Ch{
		coll: client.Database("test").Collection("users"),
		log:  log,
	}
}
