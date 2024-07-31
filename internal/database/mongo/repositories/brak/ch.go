package brak

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Ch struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func (c *Ch) Update(filter interface{}, update interface{}) error {
	_, err := c.coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}

func (c *Ch) FindByUserID(id int64) (*Brak, error) {
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

func (c *Ch) FindByKidID(id int64) (*Brak, error) {
	result := &Brak{}
	filter := bson.D{{"baby_user_id", id}}
	err := c.coll.FindOne(context.TODO(), filter).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Ch) Insert(brak *Brak) error {
	_, err := c.coll.InsertOne(context.TODO(), brak)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func (c *Ch) Delete(id primitive.ObjectID) error {
	_, err := c.coll.DeleteOne(context.TODO(), bson.D{{"_id", id}})
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func (c *Ch) FindBraksByPage(page int64, limit int64, filter interface{}) ([]*UsersBrak, int64, error) {
	var braks []*UsersBrak
	skip := (page - 1) * limit
	brakCount, err := c.coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		c.log.Sugar().Error(err)
		return nil, 0, err
	}

	pipeline := mongo.Pipeline{
		{{"$match", filter}},
		{{"$sort", bson.M{"score": -1}}},
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
	for b := range braks {
		c.log.Info("brak", zap.Any("brak", braks[b]))
	}
	return braks, brakCount, nil
}

func (c *Ch) Count(id int64) (int64, error) {
	count, err := c.coll.CountDocuments(context.TODO(),
		bson.M{"$or": []interface{}{
			bson.M{"first_user_id": id},
			bson.M{"second_user_id": id},
		}},
	)
	if err != nil {
		c.log.Sugar().Error(err)
		return 0, err
	}
	return count, nil
}

func New(client *mongo.Client, log *zap.Logger) *Ch {
	return &Ch{
		coll: client.Database("test").Collection("braks"),
		log:  log,
	}
}
