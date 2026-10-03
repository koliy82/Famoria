package media_cache

// Repository stores Telegram file identifiers for already-uploaded media, keyed
// by normalized source URL. It never stores media bytes.
type Repository interface {
	// Get returns the cached entry for a normalized URL key, or nil when absent.
	Get(key string) (*MediaCache, error)

	// Resolve returns the entry for a key, transparently following one alias hop
	// so that a short link (regenerated on every share) finds the media stored
	// under the canonical content key.
	Resolve(key string) (*MediaCache, string, error)

	// Put upserts an entry for a normalized URL key.
	Put(entry *MediaCache) error

	// PutAlias points a short-link key at the canonical key holding the media.
	PutAlias(aliasKey, canonicalKey string) error

	// Delete removes an entry, used when Telegram reports a stale file_id.
	Delete(key string) error
}
