package chat_settings

import (
	"context"
	"famoria/internal/config"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var _ Repository = (*Mongo)(nil)

type Mongo struct {
	coll  *mongo.Collection
	mu    sync.RWMutex
	cache map[int64]*ChatSettings
	log   *zap.Logger
}

func (m *Mongo) ActualData() {
	cursor, err := m.coll.Find(context.TODO(), bson.D{})
	if err != nil {
		m.log.Sugar().Error(err)
		return
	}
	var result []*ChatSettings
	if err = cursor.All(context.Background(), &result); err != nil {
		m.log.Sugar().Error(err)
		return
	}
	m.mu.Lock()
	for _, s := range result {
		m.cache[s.ChatID] = s
	}
	m.mu.Unlock()
}

func (m *Mongo) Get(chatID int64) *ChatSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.cache[chatID]; ok {
		// Return a copy so callers can't mutate the cached value.
		cp := *s
		return &cp
	}
	return &ChatSettings{ChatID: chatID, VideoConverterEnabled: false}
}

func (m *Mongo) IsVideoConverterEnabled(chatID int64) bool {
	return m.Get(chatID).VideoConverterEnabled
}

func (m *Mongo) SetVideoConverter(chatID int64, enabled bool) {
	m.mu.Lock()
	if s, ok := m.cache[chatID]; ok {
		s.VideoConverterEnabled = enabled
	} else {
		m.cache[chatID] = &ChatSettings{ChatID: chatID, VideoConverterEnabled: enabled}
	}
	m.mu.Unlock()

	_, err := m.coll.UpdateOne(
		context.TODO(),
		bson.M{"chat_id": chatID},
		bson.M{"$set": bson.M{"chat_id": chatID, "video_converter_enabled": enabled}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		m.log.Sugar().Error(err)
	}
}

func New(client *mongo.Client, log *zap.Logger, cfg config.Config) *Mongo {
	m := &Mongo{
		coll:  client.Database(cfg.MongoDatabase).Collection("chat_settings"),
		cache: make(map[int64]*ChatSettings),
		log:   log,
	}
	m.ActualData()
	return m
}
