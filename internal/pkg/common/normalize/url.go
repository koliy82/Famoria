// Package normalize turns arbitrary video-host URLs into stable cache keys.
//
// The same video can arrive under many surface forms: a youtu.be short link, a
// full watch URL with tracking parameters, a Shorts path, or a TikTok share
// link with ?_r=1&_t=... suffixes. All of them must collapse to one key, or the
// cache never hits and every repost re-downloads the file.
package normalize

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// youtubeIDPatterns match the video ID embedded in the path, across the URL
// shapes YouTube uses. The ?v= form is handled separately because there the ID
// lives in the query.
var youtubeIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(?:www\.|m\.)?youtu\.be/([A-Za-z0-9_-]{6,})$`),
	regexp.MustCompile(`^(?:www\.|m\.)?youtube\.com/shorts/([A-Za-z0-9_-]{6,})$`),
	regexp.MustCompile(`^(?:www\.|m\.)?youtube\.com/embed/([A-Za-z0-9_-]{6,})$`),
	regexp.MustCompile(`^(?:www\.|m\.)?youtube\.com/live/([A-Za-z0-9_-]{6,})$`),
	regexp.MustCompile(`^(?:www\.|m\.)?youtube\.com/v/([A-Za-z0-9_-]{6,})$`),
}

// tiktokPostPatterns match the post ID for both /video/ and /photo/ posts.
var tiktokPostPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(?:www\.)?tiktok\.com/@[\w.\-]+/(?:video|photo)/(\d+)$`),
	regexp.MustCompile(`^(?:www\.)?tiktok\.com/(?:t|embed)/([\w-]+)$`),
}

// tiktokShareHost matches vt./vm. short links, whose code identifies the target
// post directly.
var tiktokShareHost = regexp.MustCompile(`^(?:vt|vm)\.tiktok\.com/([A-Za-z0-9]+)$`)

// instagramReel matches instagram.com/reel/ID and /reels/ID.
var instagramReel = regexp.MustCompile(`^(?:www\.)?instagram\.com/reels?/([A-Za-z0-9_-]+)$`)

// trackingParams are query keys that carry no content identity: share
// attribution and client hints. Keeping them would split one video across many
// cache keys.
var trackingParams = map[string]bool{
	"si": true, "feature": true, "app": true, "utm_source": true,
	"utm_medium": true, "utm_campaign": true, "utm_term": true, "utm_content": true,
	"_r": true, "_t": true, "share_app_id": true, "share_author_id": true,
	"share_item_id": true, "share_link_id": true, "timestamp": true,
	"referer": true, "referrer": true, "fbclid": true, "gclid": true,
	"igsh": true, "igshid": true, "is_copy_url": true, "is_from_webapp": true,
	"sender_device": true, "sender_web_id": true, "checksum": true,
}

// URL converts a raw link into a stable cache key.
//
// Known hosts are reduced to their canonical content identifier
// ("youtube:dQw4w9WgXcQ", "tiktok:7633745141773257991"), so the key is
// independent of the link shape the user pasted. Unknown hosts fall back to a
// scheme-less host+path with tracking parameters stripped.
//
// An empty string is returned when the input is not an absolute URL; callers
// must treat that as "do not cache".
func URL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Links pasted without a scheme (e.g. "youtu.be/abc") are common, and
	// url.Parse would treat them as an opaque path rather than a host.
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	// Hostname strips any port and the brackets of an IPv6 literal, which manual
	// string surgery would get wrong.
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return ""
	}

	path := strings.TrimSuffix(u.Path, "/")

	if key, ok := knownHost(host, path, u.Query()); ok {
		return key
	}
	return genericKey(host, path, u.Query())
}

// knownHost reduces a recognized video host to its content identifier.
func knownHost(host, path string, q url.Values) (string, bool) {
	full := host + path

	if isYouTubeHost(host) {
		if id := youtubeID(host, path, full, q); id != "" {
			return "youtube:" + id, true
		}
	}

	if strings.HasSuffix(host, "tiktok.com") {
		if m := tiktokShareHost.FindStringSubmatch(full); m != nil {
			return "tiktok:" + m[1], true
		}
		for _, re := range tiktokPostPatterns {
			if m := re.FindStringSubmatch(full); m != nil {
				return "tiktok:" + m[1], true
			}
		}
	}

	if m := instagramReel.FindStringSubmatch(full); m != nil {
		return "instagram:" + m[1], true
	}

	return "", false
}

func isYouTubeHost(host string) bool {
	switch host {
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com",
		"youtu.be", "www.youtu.be":
		return true
	}
	return strings.HasSuffix(host, ".youtube.com")
}

// youtubeID extracts the video ID from any supported YouTube URL shape.
func youtubeID(host, path, full string, q url.Values) string {
	// The watch form keeps the ID in the query, not the path.
	if strings.EqualFold(path, "/watch") {
		return q.Get("v")
	}
	for _, re := range youtubeIDPatterns {
		if m := re.FindStringSubmatch(full); m != nil {
			return m[1]
		}
	}
	return ""
}

// genericKey is the fallback for unrecognized hosts. It keeps host+path, where
// the content identity usually lives, and drops tracking-only query parameters
// so that re-sharing a link with a fresh ?si=... suffix still hits the cache.
func genericKey(host, path string, q url.Values) string {
	keys := make([]string, 0, len(q))
	for k := range q {
		if trackingParams[strings.ToLower(k)] {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var kept []string
	for _, k := range keys {
		vals := append([]string(nil), q[k]...)
		sort.Strings(vals)
		for _, v := range vals {
			kept = append(kept, k+"="+v)
		}
	}

	key := host + path
	if len(kept) > 0 {
		key += "?" + strings.Join(kept, "&")
	}
	return "url:" + key
}
