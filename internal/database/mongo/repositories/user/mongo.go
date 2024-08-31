package user

import (
	"context"
	"errors"
	"famoria/internal/config"
	"famoria/internal/pkg/score"
	"github.com/mymmrac/telego"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"strconv"
)

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
		Score: score.Score{
			Mantissa: 0,
			Exponent: 0,
		},
	}
	if errors.Is(err, mongo.ErrNoDocuments) && actual == nil {
		model.OID = primitive.NewObjectID()
		err := c.Insert(model)
		if err != nil {
			c.log.Sugar().Error(err)
			return nil, err
		}
		c.log.Sugar().Info("Insert new user:", model)
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
		c.log.Sugar().Info("User updated: ", user)
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
	if cfg.TransferMongoDatabase != nil {
		err = TransferUsers(client, m, cfg)
		if err != nil {
			log.Sugar().Error(err)
			panic(err)
		}
	}
	return m
}

type TransferUser struct {
	OID          primitive.ObjectID `bson:"_id"`
	ID           int64              `bson:"id"`
	FirstName    string             `bson:"first_name"`
	LastName     *string            `bson:"last_name"`
	Username     *string            `bson:"username"`
	LanguageCode string             `bson:"language_code"`
	IsAdmin      *bool              `bson:"is_admin"`
}

func TransferUsers(client *mongo.Client, m *Mongo, cfg config.Config) error {
	transferColl := client.Database(*cfg.TransferMongoDatabase).Collection("users")
	usersCount, err := m.coll.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		return err
	}
	if usersCount != 0 {
		m.log.Error("transfer error: braks collection in new db is not empty")
		return nil
	}

	var transferUsers []TransferUser
	cursor, err := transferColl.Find(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	err = cursor.All(context.Background(), &transferUsers)
	if err != nil {
		return err
	}

	newUsers := make([]interface{}, len(transferUsers))
	m.log.Sugar().Info("Transfer users count: ", strconv.Itoa(len(transferUsers)))
	for i := range transferUsers {
		user := User{
			OID:          transferUsers[i].OID,
			ID:           transferUsers[i].ID,
			FirstName:    transferUsers[i].FirstName,
			LastName:     transferUsers[i].LastName,
			Username:     transferUsers[i].Username,
			LanguageCode: transferUsers[i].LanguageCode,
			Score: score.Score{
				Mantissa: 0,
				Exponent: 0,
			},
			SubscribeEnd: nil,
		}

		m.log.Sugar().Debug("Transfer user: ", zap.Any("user", user))
		newUsers[i] = user
	}

	_, err = m.coll.InsertMany(context.TODO(), newUsers)
	if err != nil {
		return err
	}
	m.log.Sugar().Info(strconv.Itoa(len(newUsers)), " users successfully transferred")
	return nil
}
