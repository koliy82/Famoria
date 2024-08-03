package user

import (
	"context"
	"errors"
	"famoria/internal/config"
	"github.com/mymmrac/telego"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Mongo struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func (c *Mongo) ValidateInfo(user *telego.User) error {
	actual, err := c.FindByID(user.ID)
	model := &User{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     &user.LastName,
		Username:     &user.Username,
		LanguageCode: user.LanguageCode,
		IsAdmin:      false,
	}
	if errors.Is(err, mongo.ErrNoDocuments) && actual == nil {
		model.OID = primitive.NewObjectID()
		err := c.Insert(model)
		if err != nil {
			c.log.Sugar().Error(err)
			return err
		}
		c.log.Sugar().Info("Insert new user:", model)
		return nil
	}

	filter := bson.M{"id": actual.ID}
	if !model.IsEquals(actual) {
		_, err := c.coll.UpdateOne(context.TODO(), filter,
			bson.M{
				"$set": bson.M{
					"first_name":    user.FirstName,
					"last_name":     user.LastName,
					"username":      user.Username,
					"language_code": user.LanguageCode,
				},
			},
		)
		if err != nil {
			c.log.Sugar().Error(err)
			return err
		}
		c.log.Sugar().Info("User updated: ", user)
	}
	return nil
}

func (c *Mongo) Replace(user *User) error {
	filter := bson.M{"id": user.ID}
	_, err := c.coll.ReplaceOne(context.TODO(), filter, user)
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}
	return nil
}

func (c *Mongo) Insert(user *User) error {
	_, err := c.coll.InsertOne(context.TODO(), user)
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}
	return nil
}

func (c *Mongo) FindByID(id int64) (*User, error) {
	user := &User{}
	filter := bson.D{{"id", id}}
	err := c.coll.FindOne(context.TODO(), filter).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, err
}

func New(client *mongo.Client, log *zap.Logger, cfg config.Config) *Mongo {
	coll := client.Database(cfg.MongoDatabase).Collection("users")
	_, err := coll.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.M{"id": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Sugar().Error(err)
	}
	return &Mongo{
		coll: coll,
		log:  log,
	}
}
