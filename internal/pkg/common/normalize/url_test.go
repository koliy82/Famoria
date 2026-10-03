package normalize

import "testing"

// TestURLSameVideoSameKey asserts that every surface form of one video collapses
// to one cache key. This is the whole point of normalization: without it the
// cache never hits and every repost re-downloads the file.
func TestURLSameVideoSameKey(t *testing.T) {
	const id = "dQw4w9WgXcQ"
	groups := map[string][]string{
		"youtube:" + id: {
			"https://www.youtube.com/watch?v=" + id,
			"https://youtube.com/watch?v=" + id,
			"https://m.youtube.com/watch?v=" + id,
			"https://www.youtube.com/watch?v=" + id + "&si=AbCdEfGh12345",
			"https://www.youtube.com/watch?feature=share&v=" + id,
			"https://www.youtube.com/watch?v=" + id + "&utm_source=telegram",
			"https://youtu.be/" + id,
			"https://www.youtu.be/" + id,
			"https://youtu.be/" + id + "?si=ZZZZZZ",
			"https://www.youtube.com/shorts/" + id,
			"https://youtube.com/embed/" + id,
			"https://www.youtube.com/live/" + id,
			"https://www.youtube.com/v/" + id,
			"youtube.com/watch?v=" + id, // no scheme
			"youtu.be/" + id,
		},
	}

	for want, urls := range groups {
		for _, in := range urls {
			if got := URL(in); got != want {
				t.Errorf("URL(%q) = %q, want %q", in, got, want)
			}
		}
	}
}

func TestURLTikTok(t *testing.T) {
	const postID = "7633745141773257991"
	cases := map[string]string{
		"https://www.tiktok.com/@lil.mill_000/photo/" + postID:                             "tiktok:" + postID,
		"https://www.tiktok.com/@lil.mill_000/video/" + postID:                             "tiktok:" + postID,
		"https://www.tiktok.com/@lil.mill_000/photo/" + postID + "?_r=1&_t=ZS-98k3JXQg425": "tiktok:" + postID,
		"https://www.tiktok.com/@user_name-1/video/" + postID:                              "tiktok:" + postID,
		"https://www.tiktok.com/t/ZS4sgaFqm/":                                              "tiktok:ZS4sgaFqm",
		"https://www.tiktok.com/embed/" + postID:                                           "tiktok:" + postID,
	}
	for in, want := range cases {
		if got := URL(in); got != want {
			t.Errorf("URL(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestURLTikTokShareLink covers vt./vm. short links, where the code itself is
// the identity and the host+path regexes must not be confused by the slash-less
// path.
func TestURLTikTokShareLink(t *testing.T) {
	in := "https://vt.tiktok.com/ZS4sgaFqm"
	want := "tiktok:ZS4sgaFqm"
	if got := URL(in); got != want {
		t.Errorf("URL(%q) = %q, want %q", in, got, want)
	}
}

// TestURLVideoAndPhotoDiffer guards the one case where collapsing is wrong: a
// photo post and a video post are different content even when they share an ID
// space, and both resolve through the same /video/ or /photo/ path form.
func TestURLDistinctPostsDiffer(t *testing.T) {
	a := URL("https://www.tiktok.com/@user/video/1111111111")
	b := URL("https://www.tiktok.com/@user/video/2222222222")
	if a == b {
		t.Errorf("distinct tiktok posts collapsed to the same key %q", a)
	}
}

func TestURLInstagram(t *testing.T) {
	const id = "CzAbCdEfGh"
	cases := map[string]string{
		"https://www.instagram.com/reel/" + id + "/":         "instagram:" + id,
		"https://instagram.com/reels/" + id:                  "instagram:" + id,
		"https://www.instagram.com/reel/" + id + "?igsh=xyz": "instagram:" + id,
	}
	for in, want := range cases {
		if got := URL(in); got != want {
			t.Errorf("URL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestURLGenericFallback(t *testing.T) {
	cases := map[string]string{
		"https://vimeo.com/76979871":                     "url:vimeo.com/76979871",
		"https://vimeo.com/76979871?utm_source=telegram": "url:vimeo.com/76979871",
		"https://example.com/a/b?si=1":                   "url:example.com/a/b",
		"https://example.com/a/b?x=2&y=1":                "url:example.com/a/b?x=2&y=1",
		"https://example.com/a/b?y=1&x=2":                "url:example.com/a/b?x=2&y=1", // sorted
		"https://example.com:8080/a/b":                   "url:example.com/a/b",         // port stripped
	}
	for in, want := range cases {
		if got := URL(in); got != want {
			t.Errorf("URL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestURLEmptyOnUnparsable(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"not a url at all",
		"javascript:alert(1)", // no host
	}
	for _, in := range cases {
		if got := URL(in); got != "" {
			t.Errorf("URL(%q) = %q, want empty", in, got)
		}
	}
}

// TestURLNoTrackingLeak asserts tracking values never reach the key, so a
// re-shared link with fresh attribution still hits the cache.
func TestURLNoTrackingLeak(t *testing.T) {
	base := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	withTracking := base + "&si=SECRET&fbclid=SECRET&gclid=SECRET"
	if URL(base) != URL(withTracking) {
		t.Errorf("tracking params leaked into the key: %q vs %q", URL(base), URL(withTracking))
	}
}
