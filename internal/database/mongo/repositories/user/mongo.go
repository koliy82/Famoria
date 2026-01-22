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

var _ Repository = (*Mongo)(nil)

type Mongo struct {
	coll *mongo.Collection
	log  *zap.Logger
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

func (c *Mongo) Update(filter interface{}, update interface{}) error {
	_, err := c.coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
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

func (c *Mongo) FindOrUpdate(user *telego.User) (*User, error) {
	actual, err := c.FindByID(user.ID)
	model := &User{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     &user.LastName,
		Username:     &user.Username,
		LanguageCode: user.LanguageCode,
		Score:        0,
	}
	if errors.Is(err, mongo.ErrNoDocuments) && actual == nil {
		model.OID = primitive.NewObjectID()
		err := c.Insert(model)
		if err != nil {
			c.log.Sugar().Error(err)
			return nil, err
		}
		c.log.Sugar().Debug("Insert new user:", model)
		return model, nil
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
			return nil, err
		}
		c.log.Sugar().Debug("User updated: ", user)
	} else {
		return actual, nil
	}
	return model, nil
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
	m := &Mongo{
		coll: coll,
		log:  log,
	}

	//var list []User
	//cursor, err := client.Database(cfg.MongoDatabase).Collection("users").Find(context.Background(), bson.M{})
	//if err != nil {
	//	panic(err)
	//}
	//for cursor.Next(context.TODO()) {
	//	var b User
	//	err := cursor.Decode(&b)
	//	if err != nil {
	//		panic(err)
	//	}
	//	b.Temp = b.Score.Mantissa
	//	list = append(list, b)
	//}
	//newValue := make([]interface{}, len(list))
	//for i := range list {
	//	newValue[i] = list[i]
	//}
	//newcoll := client.Database(cfg.MongoDatabase).Collection("newusers")
	//_, err = newcoll.InsertMany(context.Background(), newValue)
	//if err != nil {
	//	panic(err)
	//}
	//update := bson.M{"$unset": bson.M{"score": ""}}
	//_, err = newcoll.UpdateMany(context.TODO(), bson.M{}, update)
	//if err != nil {
	//	panic(err)
	//}
	//update2 := bson.D{{"$rename", bson.D{{"temp", "score"}}}}
	//_, err = newcoll.UpdateMany(context.TODO(), bson.M{}, update2)
	//if err != nil {
	//	panic(err)
	//}

	return m
}
