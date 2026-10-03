package media_cache

import (
	"context"
	"errors"
	"time"

	"famoria/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var _ Repository = (*Mongo)(nil)

type Mongo struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func (c *Mongo) Get(key string) (*MediaCache, error) {
	if key == "" {
		return nil, errors.New("media_cache: empty key")
	}
	var entry MediaCache
	err := c.coll.FindOne(context.TODO(), bson.M{"_id": key}).Decode(&entry)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		c.log.Error("media_cache: lookup failed", zap.String("key", key), zap.Error(err))
		return nil, err
	}
	return &entry, nil
}

func (c *Mongo) Resolve(key string) (*MediaCache, string, error) {
	return resolveAlias(key, c.Get)
}

// resolveAlias looks a key up and follows at most one alias hop, returning the
// entry and the key it actually lives under.
//
// Returning the resolved key matters: when Telegram rejects a stale file_id the
// caller deletes the row, and deleting the alias instead of the target would
// leave the dead file_id in place to be rejected again on every later share.
//
// A single hop is deliberate — an alias pointing at another alias is reported as
// a miss rather than followed, so a cycle cannot loop here forever.
func resolveAlias(key string, get func(string) (*MediaCache, error)) (*MediaCache, string, error) {
	entry, err := get(key)
	if err != nil || entry == nil {
		return entry, key, err
	}
	if !entry.IsAlias() {
		return entry, key, nil
	}

	target := entry.AliasTarget
	canonical, err := get(target)
	if err != nil {
		return nil, target, err
	}
	if canonical == nil || canonical.IsAlias() {
		return nil, target, nil
	}
	return canonical, target, nil
}

func (c *Mongo) Put(entry *MediaCache) error {
	if entry == nil || entry.URL == "" {
		return errors.New("media_cache: entry without a URL key")
	}
	if !entry.Replayable() {
		return errors.New("media_cache: entry has nothing replayable")
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	_, err := c.coll.UpdateOne(
		context.TODO(),
		bson.M{"_id": entry.URL},
		bson.M{"$set": entry},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		c.log.Error("media_cache: upsert failed", zap.String("key", entry.URL), zap.Error(err))
		return err
	}
	return nil
}

func (c *Mongo) PutAlias(aliasKey, canonicalKey string) error {
	if aliasKey == "" || canonicalKey == "" || aliasKey == canonicalKey {
		return nil
	}

	entry := &MediaCache{
		URL:         aliasKey,
		Kind:        KindAlias,
		AliasTarget: canonicalKey,
		CreatedAt:   time.Now(),
	}
	_, err := c.coll.UpdateOne(
		context.TODO(),
		bson.M{"_id": aliasKey},
		bson.M{"$set": entry},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		c.log.Error("media_cache: alias upsert failed",
			zap.String("alias", aliasKey), zap.String("target", canonicalKey), zap.Error(err))
		return err
	}
	return nil
}

func (c *Mongo) Delete(key string) error {
	if key == "" {
		return nil
	}
	if _, err := c.coll.DeleteOne(context.TODO(), bson.M{"_id": key}); err != nil {
		c.log.Error("media_cache: delete failed", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

func New(client *mongo.Client, log *zap.Logger, cfg config.Config) *Mongo {
	return &Mongo{
		coll: client.Database(cfg.MongoDatabase).Collection("media_cache"),
		log:  log,
	}
}
