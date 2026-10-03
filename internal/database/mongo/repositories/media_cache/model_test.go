package media_cache

import (
	"errors"
	"testing"
)

func TestKindPredicates(t *testing.T) {
	video := &MediaCache{Kind: KindVideo, FileID: "abc"}
	if !video.IsVideo() || video.IsPhoto() || video.IsAlias() {
		t.Error("video entry misclassified")
	}
	if !video.Replayable() {
		t.Error("video entry should be replayable")
	}

	photo := &MediaCache{Kind: KindPhoto, FileIDs: []string{"p1", "p2"}}
	if !photo.IsPhoto() || photo.IsVideo() || photo.IsAlias() {
		t.Error("photo entry misclassified")
	}
	if !photo.Replayable() {
		t.Error("photo entry should be replayable")
	}

	alias := &MediaCache{Kind: KindAlias, AliasTarget: "tiktok:123"}
	if !alias.IsAlias() || alias.IsVideo() || alias.IsPhoto() {
		t.Error("alias entry misclassified")
	}
	// An alias carries no media of its own, so it must not be replayed directly:
	// the caller has to resolve it to its target first.
	if alias.Replayable() {
		t.Error("alias entry should not be replayable")
	}
}

// TestPredicatesRequirePayload guards against a row that names the right kind
// but carries nothing to send — replaying those would fail every time and never
// fall back to a fresh download.
func TestPredicatesRequirePayload(t *testing.T) {
	cases := []*MediaCache{
		{Kind: KindVideo},                      // no FileID
		{Kind: KindPhoto},                      // no FileIDs
		{Kind: KindPhoto, FileIDs: []string{}}, // empty album
		{Kind: KindAlias},                      // no target
		{Kind: KindAlias, AliasTarget: ""},     // empty target
	}
	for i, c := range cases {
		if c.Replayable() {
			t.Errorf("case %d: %+v should not be replayable", i, c)
		}
	}
}

// TestPutRejectsIncompleteEntries covers the validation in the Mongo layer,
// which is the last line of defence before a useless row reaches the database.
func TestPutRejectsIncompleteEntries(t *testing.T) {
	c := &Mongo{}

	if err := c.Put(nil); err == nil {
		t.Error("Put(nil) should error")
	}
	if err := c.Put(&MediaCache{}); err == nil {
		t.Error("Put without a key should error")
	}
	if err := c.Put(&MediaCache{URL: "k", Kind: KindVideo}); err == nil {
		t.Error("Put of a video without a file_id should error")
	}
}

// TestPutAliasIgnoresDegenerateKeys ensures self-referential aliases are never
// written: resolveAlias follows exactly one hop, so a key pointing at itself
// would resolve to nothing and the cache would appear broken rather than merely
// empty.
func TestPutAliasIgnoresDegenerateKeys(t *testing.T) {
	c := &Mongo{}

	for _, tc := range []struct{ alias, target string }{
		{"", "canonical"},
		{"alias", ""},
		{"same", "same"},
	} {
		if err := c.PutAlias(tc.alias, tc.target); err != nil {
			t.Errorf("PutAlias(%q, %q) should be a no-op, got %v", tc.alias, tc.target, err)
		}
	}
}

func TestDeleteEmptyKeyIsNoop(t *testing.T) {
	c := &Mongo{}
	if err := c.Delete(""); err != nil {
		t.Errorf("Delete(\"\") should be a no-op, got %v", err)
	}
}

func TestGetRejectsEmptyKey(t *testing.T) {
	c := &Mongo{}
	if _, err := c.Get(""); err == nil {
		t.Error("Get(\"\") should error rather than hit the database")
	}
}

// stubGet builds a resolver over a fixed map, so alias resolution is testable
// without a live Mongo.
func stubGet(rows map[string]*MediaCache) func(string) (*MediaCache, error) {
	return func(key string) (*MediaCache, error) {
		return rows[key], nil
	}
}

// TestResolveAliasReturnsCanonicalKey is the core guarantee behind alias
// storage: a short link resolves to the entry stored under the canonical key,
// and the returned key is the canonical one so a later delete targets the real
// row rather than the alias.
func TestResolveAliasReturnsCanonicalKey(t *testing.T) {
	canonical := &MediaCache{URL: "tiktok:123", Kind: KindVideo, FileID: "fid"}
	rows := map[string]*MediaCache{
		"url:vt.tiktok.com/abc": {URL: "url:vt.tiktok.com/abc", Kind: KindAlias, AliasTarget: "tiktok:123"},
		"tiktok:123":            canonical,
	}

	entry, key, err := resolveAlias("url:vt.tiktok.com/abc", stubGet(rows))
	if err != nil {
		t.Fatal(err)
	}
	if entry != canonical {
		t.Error("expected the canonical entry to be returned")
	}
	if key != "tiktok:123" {
		t.Errorf("resolved key = %q, want the canonical %q", key, "tiktok:123")
	}
}

// TestResolveAliasDirectHit verifies a non-alias key is returned unchanged, so
// ordinary canonical links do not pay for a second lookup.
func TestResolveAliasDirectHit(t *testing.T) {
	video := &MediaCache{URL: "youtube:abc", Kind: KindVideo, FileID: "fid"}
	entry, key, err := resolveAlias("youtube:abc", stubGet(map[string]*MediaCache{"youtube:abc": video}))
	if err != nil {
		t.Fatal(err)
	}
	if entry != video || key != "youtube:abc" {
		t.Errorf("direct hit should return the entry unchanged, got key=%q", key)
	}
}

// TestResolveAliasMissingTarget verifies an alias whose target was never cached
// is a clean miss, not an error — the caller then downloads normally.
func TestResolveAliasMissingTarget(t *testing.T) {
	rows := map[string]*MediaCache{
		"url:vt.tiktok.com/x": {URL: "url:vt.tiktok.com/x", Kind: KindAlias, AliasTarget: "tiktok:gone"},
	}
	entry, key, err := resolveAlias("url:vt.tiktok.com/x", stubGet(rows))
	if err != nil {
		t.Fatalf("a missing target is a miss, not an error: %v", err)
	}
	if entry != nil {
		t.Error("expected a nil entry for an alias with no cached target")
	}
	if key != "tiktok:gone" {
		t.Errorf("resolved key = %q, want the target %q", key, "tiktok:gone")
	}
}

// TestResolveAliasStopsAtOneHop guards the no-cycle guarantee: an alias pointing
// at another alias resolves to a miss instead of following the chain, so a
// corrupt or cyclic row cannot loop forever.
func TestResolveAliasStopsAtOneHop(t *testing.T) {
	rows := map[string]*MediaCache{
		"a": {URL: "a", Kind: KindAlias, AliasTarget: "b"},
		"b": {URL: "b", Kind: KindAlias, AliasTarget: "c"},
		"c": {URL: "c", Kind: KindVideo, FileID: "fid"},
	}
	entry, _, err := resolveAlias("a", stubGet(rows))
	if err != nil {
		t.Fatal(err)
	}
	if entry != nil {
		t.Error("resolveAlias must follow at most one hop, so a->b->c should miss")
	}
}

// TestResolveAliasPropagatesLookupError verifies a database failure surfaces
// rather than being swallowed into a false miss, which would hide a broken cache.
func TestResolveAliasPropagatesLookupError(t *testing.T) {
	boom := errors.New("db down")
	get := func(string) (*MediaCache, error) { return nil, boom }

	if _, _, err := resolveAlias("k", get); !errors.Is(err, boom) {
		t.Errorf("expected the lookup error to propagate, got %v", err)
	}

	// The same must hold for the error on the second (target) hop.
	rows := map[string]*MediaCache{
		"a": {URL: "a", Kind: KindAlias, AliasTarget: "b"},
	}
	get2 := func(key string) (*MediaCache, error) {
		if key == "a" {
			return rows["a"], nil
		}
		return nil, boom
	}
	if _, _, err := resolveAlias("a", get2); !errors.Is(err, boom) {
		t.Errorf("expected the target lookup error to propagate, got %v", err)
	}
}
