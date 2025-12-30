package brak

import (
	"context"
	"famoria/internal/bot/idle/event"
	"famoria/internal/bot/idle/event/anubis"
	"famoria/internal/bot/idle/event/casino"
	"famoria/internal/bot/idle/event/events"
	"famoria/internal/bot/idle/event/growkid"
	"famoria/internal/bot/idle/event/hamster"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/inventory"
	"famoria/internal/bot/idle/item/items"
	"famoria/internal/config"
	"famoria/internal/pkg/common"
	"math"
	"strconv"
	"time"

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
			"score.exponent": -1,
			"score.mantissa": -1,
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
	if cfg.TransferMongoDatabase != nil {
		err = TransferBraks(client, m, cfg)
		if err != nil {
			log.Sugar().Error(err)
			panic(err)
		}
	}
	return m
}

type TransferBrak struct {
	OID               primitive.ObjectID `bson:"_id"`
	FirstUserID       int64              `bson:"first_user_id"`
	SecondUserID      int64              `bson:"second_user_id"`
	ChatID            int64              `bson:"chat_id,omitempty"`
	CreateDate        time.Time          `bson:"create_date"`
	BabyUserID        *int64             `bson:"baby_user_id"`
	BabyCreateDate    *time.Time         `bson:"baby_create_date"`
	Score             int64              `bson:"score"`
	LastCasinoPlay    time.Time          `bson:"last_casino_play"`
	LastGrowKid       time.Time          `bson:"last_grow_kid"`
	LastHamsterUpdate time.Time          `bson:"last_hamster_update"`
	TapCount          int                `bson:"tap_count"`
}

func TransferBraks(client *mongo.Client, m *Mongo, cfg config.Config) error {
	transferColl := client.Database(*cfg.TransferMongoDatabase).Collection("braks")
	braksCount, err := m.coll.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		return err
	}
	if braksCount != 0 {
		m.log.Warn("transfer brak collection in new db is not empty, skip transfer")
		return nil
	}

	var transferBraks []TransferBrak
	cursor, err := transferColl.Find(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	err = cursor.All(context.TODO(), &transferBraks)
	if err != nil {
		return err
	}

	newBraks := make([]interface{}, len(transferBraks))
	m.log.Sugar().Info("Transfer braks count: ", strconv.Itoa(len(transferBraks)))
	for i := range transferBraks {
		brak := Brak{
			OID:            transferBraks[i].OID,
			FirstUserID:    transferBraks[i].FirstUserID,
			SecondUserID:   transferBraks[i].SecondUserID,
			ChatID:         transferBraks[i].ChatID,
			CreateDate:     transferBraks[i].CreateDate,
			BabyUserID:     transferBraks[i].BabyUserID,
			BabyCreateDate: transferBraks[i].BabyCreateDate,
			Score:          &common.Score{Mantissa: transferBraks[i].Score},
			Inventory:      &inventory.Inventory{Items: make(map[items.ItemId]inventory.Item)},
			Events: &events.Events{
				Hamster: &hamster.Hamster{
					Base: event.Base{
						LastPlay:  transferBraks[i].LastHamsterUpdate,
						PlayCount: uint16(transferBraks[i].TapCount),
					},
				},
				Casino: &casino.Casino{
					Base: event.Base{
						LastPlay:  transferBraks[i].LastCasinoPlay,
						PlayCount: 1,
					},
				},
				GrowKid: &growkid.GrowKid{
					Base: event.Base{
						LastPlay:  transferBraks[i].LastGrowKid,
						PlayCount: 1,
					},
				},
				Anubis: &anubis.Anubis{
					Base: event.Base{
						LastPlay:  time.Time{},
						PlayCount: 0,
					},
				},
			},
			SubscribeEnd: nil,
		}
		if brak.Score.Mantissa < 0 {
			brak.Score.Mantissa = int64(math.Abs(float64(brak.Score.Mantissa)))
		}

		m.log.Sugar().Debug("transfer brak: ", zap.Any("brak", transferBraks[i]))
		newBraks[i] = brak
	}

	_, err = m.coll.InsertMany(context.TODO(), newBraks)
	if err != nil && len(transferBraks) != 0 {
		return err
	}
	m.log.Sugar().Info(strconv.Itoa(len(newBraks)), " braks successfully transferred")
	return nil
}
