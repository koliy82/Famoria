package brak

import (
	"context"
	"famoria/internal/bot/idle/item"
	"famoria/internal/config"

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

func (c *Mongo) Update(filter interface{}, update interface{}) error {
	_, err := c.coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}

func (c *Mongo) FindByUserID(id int64, m *item.Manager) (*Brak, error) {
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
	if m != nil {
		result.ApplyBuffs(m)
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
		{{"$sort", bson.M{
			"score": -1,
		}}},
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
				"_id":              "$_id",
				"first_user_id":    "$first_user_id",
				"second_user_id":   "$second_user_id",
				"chat_id":          "$chat_id",
				"create_date":      "$create_date",
				"baby_user_id":     "$baby_user_id",
				"baby_create_date": "$baby_create_date",
				"score":            "$score",
				"subscribe_end":    "$subscribe_end",
				//"inventory":        "$inventory",
				//"events":           "$events",
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

func (c *Mongo) FindAllMining() ([]*Brak, error) {
	var braks []*Brak
	filter := bson.M{"events.mining": bson.M{"$exists": true}}
	cursor, err := c.coll.Find(context.TODO(), filter)
	if err != nil {
		c.log.Sugar().Error(err)
		return nil, err
	}
	err = cursor.All(context.TODO(), &braks)
	if err != nil {
		c.log.Sugar().Error(err)
		return nil, err
	}
	return braks, nil
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
	m := &Mongo{
		coll: coll,
		log:  log,
	}

	//var list []Brak
	//cursor, err := client.Database(cfg.MongoDatabase).Collection("braks").Find(context.Background(), bson.M{})
	//if err != nil {
	//	panic(err)
	//}
	//for cursor.Next(context.TODO()) {
	//	var b Brak
	//	err := cursor.Decode(&b)
	//	if err != nil {
	//		panic(err)
	//	}
	//	b.Temp = b.Score.Mantissa
	//	b.Score = nil
	//	list = append(list, b)
	//}
	//newValue := make([]interface{}, len(list))
	//for i := range list {
	//	newValue[i] = list[i]
	//}
	//newcoll := client.Database(cfg.MongoDatabase).Collection("newbraks")
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
