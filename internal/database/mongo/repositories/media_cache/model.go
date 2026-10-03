package media_cache

import "time"

// Kind distinguishes cached video files from cached photo albums, because they
// are replayed through different Bot API methods.
type Kind string

const (
	KindVideo Kind = "video"
	KindPhoto Kind = "photo"
	// KindAlias is a pointer row, not media. It exists because short links
	// (vt.tiktok.com/X, youtu.be with a fresh ?si=) are regenerated on every
	// share, so mapping them to the canonical content key is what lets the cache
	// hit at all instead of re-downloading the same video every time.
	KindAlias Kind = "alias"
)

// MediaCache is one cached upload, keyed by the normalized source URL.
//
// Only Telegram-side identifiers and small metadata are stored — never the
// media bytes. The file lives on Telegram's servers and is replayed via its
// file_id, which is why disk space on the bot host stays constant.
type MediaCache struct {
	// URL is the normalized cache key, e.g. "youtube:dQw4w9WgXcQ".
	URL string `bson:"_id"`

	// Kind selects the replay method: a single video, or a photo album.
	Kind Kind `bson:"kind"`

	// AliasTarget is the canonical cache key this row points at. Set only when
	// Kind is KindAlias, in which case the media fields below are empty.
	AliasTarget string `bson:"alias_target,omitempty"`

	// FileID is the Telegram identifier of the uploaded video. Set for KindVideo.
	FileID string `bson:"file_id,omitempty"`

	// FileIDs holds the Telegram identifiers of every uploaded photo, in album
	// order. Set for KindPhoto.
	FileIDs []string `bson:"file_ids,omitempty"`

	// Title and Uploader are kept only for logging and debugging; they are cheap
	// and make a cache row self-describing.
	Title    string `bson:"title,omitempty"`
	Uploader string `bson:"uploader,omitempty"`

	// Duration, Width and Height are the video geometry captured at upload time.
	// Telegram does not recompute them when replaying by file_id, so passing the
	// original values keeps the cached send looking identical to a fresh upload.
	Duration int `bson:"duration,omitempty"`
	Width    int `bson:"width,omitempty"`
	Height   int `bson:"height,omitempty"`

	CreatedAt time.Time `bson:"created_at"`
}

// IsVideo reports whether this entry replays as a single video.
func (m *MediaCache) IsVideo() bool { return m.Kind == KindVideo && m.FileID != "" }

// IsPhoto reports whether this entry replays as a photo or photo album.
func (m *MediaCache) IsPhoto() bool { return m.Kind == KindPhoto && len(m.FileIDs) > 0 }

// IsAlias reports whether this entry merely points at a canonical key.
func (m *MediaCache) IsAlias() bool { return m.Kind == KindAlias && m.AliasTarget != "" }

// Replayable reports whether the entry carries enough to resend media directly.
// Aliases are deliberately excluded: they hold no file identifiers, so a caller
// must resolve them to their target first.
func (m *MediaCache) Replayable() bool { return m.IsVideo() || m.IsPhoto() }
