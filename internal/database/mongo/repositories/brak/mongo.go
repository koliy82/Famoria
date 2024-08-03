package brak

import (
	"context"
	"errors"
	"famoria/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type Mongo struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func (c *Mongo) Update(filter interface{}, update interface{}) error {
	_, err := c.coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}

func (c *Mongo) FindByUserID(id int64) (*Brak, error) {
	result := &Brak{}
	filter := bson.D{
		{"$or", []interface{}{
			bson.D{{"first_user_id", id}},
			bson.D{{"second_user_id", id}},
		}},
	}
	err := c.coll.FindOne(context.TODO(), filter).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Mongo) FindByKidID(id int64) (*Brak, error) {
	result := &Brak{}
	filter := bson.D{{"baby_user_id", id}}
	err := c.coll.FindOne(context.TODO(), filter).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Mongo) Insert(brak *Brak) error {
	_, err := c.coll.InsertOne(context.TODO(), brak)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func (c *Mongo) Delete(id primitive.ObjectID) error {
	_, err := c.coll.DeleteOne(context.TODO(), bson.D{{"_id", id}})
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func (c *Mongo) FindBraksByPage(page int64, limit int64, filter interface{}) ([]*UsersBrak, int64, error) {
	var braks []*UsersBrak
	skip := (page - 1) * limit
	brakCount, err := c.coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		c.log.Sugar().Error(err)
		return nil, 0, err
	}
	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$sort", bson.M{"score": -1, "create_date": -1}}},
		{{"$skip", skip}},
		{{"$limit", limit}},
		{{"$lookup", bson.M{
			"from":         "users",
			"localField":   "first_user_id",
			"foreignField": "id",
			"as":           "first",
		}}},
		{{"$lookup", bson.M{
			"from":         "users",
			"localField":   "second_user_id",
			"foreignField": "id",
			"as":           "second",
		},
		}},
		{{"$lookup", bson.M{
			"from":         "users",
			"localField":   "baby_user_id",
			"foreignField": "id",
			"as":           "baby",
		}},
		},
		{{"$unwind", bson.M{
			"path":                       "$first",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{"$unwind", bson.M{
			"path":                       "$second",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{"$unwind", bson.M{
			"path":                       "$baby",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{"$project", bson.M{
			"brak": bson.M{
				"_id":                 "$_id",
				"first_user_id":       "$first_user_id",
				"second_user_id":      "$second_user_id",
				"chat_id":             "$chat_id",
				"create_date":         "$create_date",
				"baby_user_id":        "$baby_user_id",
				"baby_create_date":    "$baby_create_date",
				"score":               "$score",
				"last_casino_play":    "$last_casino_play",
				"last_grow_kid":       "$last_grow_kid",
				"last_hamster_update": "$last_hamster_update",
				"tap_count":           "$tap_count",
			},
			"first":  1,
			"second": 1,
			"baby":   1,
		}},
		},
	}

	cursor, err := c.coll.Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.log.Sugar().Error(err)
		return nil, 0, err
	}
	err = cursor.All(context.TODO(), &braks)
	if err != nil {
		c.log.Sugar().Error(err)
		return nil, 0, err
	}
	return braks, brakCount, nil
}

func (c *Mongo) Count(filter interface{}) (int64, error) {
	count, err := c.coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		c.log.Sugar().Error(err)
		return 0, err
	}
	return count, nil
}

func New(client *mongo.Client, log *zap.Logger, cfg config.Config) *Mongo {
	coll := client.Database(cfg.MongoDatabase).Collection("braks")
	_, err := coll.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{"first_user_id", 1},
				{"second_user_id", 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"baby_user_id": 1},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{"baby_user_id": bson.M{"$exists": true}}),
		},
	})
	if err != nil {
		log.Sugar().Error(err)
	}
	_ = TransferBraks(client, log, cfg)
	return &Mongo{
		coll: coll,
		log:  log,
	}
}

type TransferBrak struct {
	OID          primitive.ObjectID `bson:"_id"`
	FirstUserID  int64              `bson:"firstUserID"`
	SecondUserID int64              `bson:"secondUserID"`
	CreateDate   int64              `bson:"time"`
	Baby         *TransferBrakBaby  `bson:"baby"`
}

type TransferBrakBaby struct {
	BabyUserID int64 `bson:"userID"`
	time       int64 `bson:"time"`
}

func TransferBraks(client *mongo.Client, log *zap.Logger, cfg config.Config) error {
	transferColl := client.Database("aratossik").Collection("braks")
	coll := client.Database(cfg.MongoDatabase).Collection("braks")
	braksCount, _ := coll.CountDocuments(context.TODO(), bson.D{})

	if braksCount != 0 {
		return errors.New("braks already exists")
	}

	var transferBraks []TransferBrak
	cursor, err := transferColl.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Sugar().Error(err)
		return nil
	}

	err = cursor.All(context.TODO(), &transferBraks)
	if err != nil {
		log.Sugar().Error(err)
		return nil
	}

	newBraks := make([]interface{}, len(transferBraks))
	log.Info(strconv.Itoa(len(transferBraks)))
	for i := range transferBraks {
		brak := Brak{
			OID:          transferBraks[i].OID,
			FirstUserID:  transferBraks[i].FirstUserID,
			SecondUserID: transferBraks[i].SecondUserID,
			ChatID:       0,
			CreateDate:   time.UnixMilli(transferBraks[i].CreateDate),
			Score:        0,
			TapCount:     50,
		}

		log.Sugar().Info("transfer brak: ", zap.Any("brak", transferBraks[i]))
		if transferBraks[i].Baby != nil {
			brak.BabyUserID = &transferBraks[i].Baby.BabyUserID
			date := time.Unix(transferBraks[i].Baby.time, 0)
			brak.BabyCreateDate = &date
			log.Sugar().Info("transfer baby: ", zap.Any("baby", transferBraks[i].Baby.BabyUserID))
		}
		newBraks[i] = brak
	}

	_, err = coll.InsertMany(context.TODO(), newBraks)
	if err != nil && len(transferBraks) != 0 {
		log.Sugar().Error(err)
		panic(err)
		return nil
	}

	return nil
}
