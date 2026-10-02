package chat_settings

import (
	"context"
	"sync"

	"famoria/internal/config"
	"famoria/internal/pkg/i18n"

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
		if s.DisabledFeatures == nil {
			s.DisabledFeatures = make(map[string]bool)
		}
		m.cache[s.ChatID] = s
	}
	m.mu.Unlock()
}

// getOrCreateLocked returns the cached settings for a chat, creating (but not
// persisting) defaults when absent. The caller must hold the write lock.
func (m *Mongo) getOrCreateLocked(chatID int64) *ChatSettings {
	if s, ok := m.cache[chatID]; ok {
		if s.DisabledFeatures == nil {
			s.DisabledFeatures = make(map[string]bool)
		}
		return s
	}
	s := newDefault(chatID)
	m.cache[chatID] = s
	return s
}

// persistLocked writes the given settings back to Mongo as a whole document.
// Upserting the full document avoids partial updates drifting from the cache.
// The caller must hold the write lock.
func (m *Mongo) persistLocked(s *ChatSettings) {
	_, err := m.coll.UpdateOne(
		context.TODO(),
		bson.M{"chat_id": s.ChatID},
		bson.M{"$set": s},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		m.log.Sugar().Error(err)
	}
}

func (m *Mongo) Get(chatID int64) *ChatSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.cache[chatID]; ok {
		// Return a copy so callers can't mutate the cached value.
		return s.clone()
	}
	return newDefault(chatID)
}

func (m *Mongo) Lang(chatID int64) i18n.Lang {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.cache[chatID]; ok {
		return s.Language()
	}
	return i18n.DefaultLang
}

func (m *Mongo) IsVideoConverterEnabled(chatID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.cache[chatID]; ok {
		return s.VideoConverterEnabled
	}
	return false
}

func (m *Mongo) DeleteOriginalLink(chatID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.cache[chatID]; ok {
		return s.DeleteOriginalLink()
	}
	return true
}

func (m *Mongo) IsFeatureEnabled(chatID int64, f Feature) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.cache[chatID]; ok {
		return s.IsFeatureEnabled(f)
	}
	// No stored settings: every feature is on.
	return true
}

func (m *Mongo) SetLang(chatID int64, lang i18n.Lang) {
	if !lang.IsValid() {
		lang = i18n.DefaultLang
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.getOrCreateLocked(chatID)
	s.Lang = string(lang)
	m.persistLocked(s)
}

func (m *Mongo) SetVideoConverter(chatID int64, enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.getOrCreateLocked(chatID)
	s.VideoConverterEnabled = enabled
	m.persistLocked(s)
}

func (m *Mongo) SetVideoKeepOriginal(chatID int64, keep bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.getOrCreateLocked(chatID)
	s.VideoKeepOriginal = keep
	m.persistLocked(s)
}

func (m *Mongo) SetFeatureEnabled(chatID int64, f Feature, enabled bool) {
	normalized := normalizeFeature(string(f))
	if normalized == "" {
		m.log.Warn("chat_settings: ignoring unknown feature",
			zap.Int64("chat_id", chatID), zap.String("feature", string(f)))
		return
	}
	f = normalized

	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.getOrCreateLocked(chatID)
	if enabled {
		delete(s.DisabledFeatures, string(f))
	} else {
		s.DisabledFeatures[string(f)] = true
	}
	m.persistLocked(s)
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
