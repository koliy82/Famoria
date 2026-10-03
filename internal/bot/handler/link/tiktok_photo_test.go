package link

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// noAmbientProxy clears proxy environment variables for the duration of a test.
//
// proxyClient with an empty specification returns http.DefaultClient, which
// honours HTTP_PROXY/HTTPS_PROXY/ALL_PROXY. On a machine where those are set, an
// assertion about an unreachable host would instead route through a real proxy and
// pass or fail depending on the environment rather than the code.
func noAmbientProxy(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"HTTP_PROXY", "http_proxy",
		"HTTPS_PROXY", "https_proxy",
		"ALL_PROXY", "all_proxy",
		"CGI_HTTP_PROXY", "cgi_http_proxy",
	} {
		t.Setenv(k, "")
	}
}

func TestIsTikTokShortLink(t *testing.T) {
	cases := map[string]bool{
		"https://vt.tiktok.com/ZSbmeYhHC/":     true,
		"https://vm.tiktok.com/ZSbmeYhHC":      true,
		"http://vt.tiktok.com/abc":             true,
		"https://www.tiktok.com/@u/video/123":  false,
		"https://www.tiktok.com/@u/photo/123":  false,
		"https://vt.tiktok.com.evil.example/x": false,
		"https://example.com/vt.tiktok.com/x":  false,
		"https://youtu.be/abc123456":           false,
		"":                                     false,
		"not a url":                            false,
	}
	for in, want := range cases {
		if got := isTikTokShortLink(in); got != want {
			t.Errorf("isTikTokShortLink(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsTikTokPostURL(t *testing.T) {
	cases := map[string]bool{
		"https://www.tiktok.com/@u/video/123": true,
		"https://www.tiktok.com/@u/photo/123": true,
		"https://www.tiktok.com/@u/PHOTO/123": true, // case-insensitive segment
		"https://vt.tiktok.com/ZSbmeYhHC/":    false,
		"https://www.tiktok.com/@u":           false,
		"https://youtu.be/abc":                false,
		"":                                    false,
	}
	for in, want := range cases {
		if got := isTikTokPostURL(in); got != want {
			t.Errorf("isTikTokPostURL(%q) = %v, want %v", in, got, want)
		}
	}
}

// TestIsTikTokPhotoPostRequiresResolvedURL documents the bug this fixes: a share
// link carries no /photo/ segment, so it must NOT be classified as a photo post.
// Sending one down the video path made yt-dlp answer "Unsupported URL".
func TestIsTikTokPhotoPostRequiresResolvedURL(t *testing.T) {
	if isTikTokPhotoPost("https://vt.tiktok.com/ZSbmeYhHC/") {
		t.Error("a share link must not be classified as a photo post before resolving")
	}
	if !isTikTokPhotoPost("https://www.tiktok.com/@lil.mill_000/photo/7633745141773257991") {
		t.Error("a resolved /photo/ URL must be classified as a photo post")
	}
	if isTikTokPhotoPost("https://www.tiktok.com/@u/video/7633745141773257991") {
		t.Error("a /video/ URL must not be classified as a photo post")
	}
}

// TestVideoURLFromPhotoPreservesUsernameCase guards the earlier regression:
// lowercasing the whole path broke usernames containing capitals.
func TestVideoURLFromPhotoPreservesUsernameCase(t *testing.T) {
	got := videoURLFromPhoto("https://www.tiktok.com/@UserName/photo/123?_r=1")
	if !strings.Contains(got, "@UserName") {
		t.Errorf("username case was mangled: %q", got)
	}
	if !strings.Contains(got, "/video/123") {
		t.Errorf("path segment not rewritten: %q", got)
	}
	if strings.Contains(got, "_r=1") {
		t.Errorf("query should be stripped: %q", got)
	}
}

func TestFetchURLFallsBackToOriginal(t *testing.T) {
	r := sendRequest{originalURL: "https://example.com/a"}
	if got := r.fetchURL(); got != "https://example.com/a" {
		t.Errorf("fetchURL() = %q, want the original when mediaURL is unset", got)
	}

	r2 := r.withMediaURL("https://example.com/canonical")
	if got := r2.fetchURL(); got != "https://example.com/canonical" {
		t.Errorf("fetchURL() = %q, want the resolved URL", got)
	}
	// The original must survive, since the caption links back to what the user
	// actually pasted.
	if r2.originalURL != "https://example.com/a" {
		t.Errorf("withMediaURL clobbered originalURL: %q", r2.originalURL)
	}
}

// TestWithMediaURLReturnsCopy ensures the receiver is not mutated, so a caller
// that keeps using the original sendRequest is unaffected.
func TestWithMediaURLReturnsCopy(t *testing.T) {
	r := sendRequest{originalURL: "https://example.com/a"}
	_ = r.withMediaURL("https://example.com/b")
	if r.mediaURL != "" {
		t.Errorf("withMediaURL mutated the receiver: mediaURL=%q", r.mediaURL)
	}
}

// --- redirect resolution -------------------------------------------------

// newRedirectServer serves a share link that redirects to the given target,
// optionally through an intermediate hop.
func newRedirectServer(t *testing.T, target string, hops int) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/short", func(w http.ResponseWriter, r *http.Request) {
		if hops > 0 {
			hops--
			http.Redirect(w, r, "/intermediate", http.StatusFound)
			return
		}
		http.Redirect(w, r, target, http.StatusFound)
	})
	mux.HandleFunc("/intermediate", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
	})
	return httptest.NewServer(mux)
}

func TestResolveShortLinkFollowsToPhotoPost(t *testing.T) {
	noAmbientProxy(t)
	const post = "https://www.tiktok.com/@lil.mill_000/photo/7633745141773257991?_r=1&_t=X"

	srv := newRedirectServer(t, post, 0)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	got, ok := resolveTikTokShortLink(ctx, srv.URL+"/short", "")
	if !ok {
		t.Fatal("resolution reported failure")
	}
	if got != post {
		t.Errorf("resolved = %q, want %q", got, post)
	}
	// The whole point of resolving: the result must now classify as a photo post.
	if !isTikTokPhotoPost(got) {
		t.Errorf("resolved URL %q is not classified as a photo post", got)
	}
}

func TestResolveShortLinkFollowsToVideoPost(t *testing.T) {
	noAmbientProxy(t)
	const post = "https://www.tiktok.com/@ahahahchto/video/7674718638846315797?_r=1"

	srv := newRedirectServer(t, post, 0)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	got, ok := resolveTikTokShortLink(ctx, srv.URL+"/short", "")
	if !ok {
		t.Fatal("resolution reported failure")
	}
	if got != post {
		t.Errorf("resolved = %q, want %q", got, post)
	}
	if isTikTokPhotoPost(got) {
		t.Errorf("a /video/ URL must not classify as a photo post: %q", got)
	}
}

// TestResolveShortLinkFollowsIntermediateHops covers multi-hop chains, which real
// share links use.
func TestResolveShortLinkFollowsIntermediateHops(t *testing.T) {
	noAmbientProxy(t)
	const post = "https://www.tiktok.com/@u/photo/123"

	srv := newRedirectServer(t, post, 2)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	got, ok := resolveTikTokShortLink(ctx, srv.URL+"/short", "")
	if !ok || got != post {
		t.Errorf("resolved = %q (ok=%v), want %q", got, ok, post)
	}
}

// TestResolveShortLinkDegradesOnUnreachable verifies the failure mode: an
// unresolvable link returns the original rather than an error, so a video post
// still downloads via yt-dlp (which handles short links itself) instead of the
// whole job failing.
func TestResolveShortLinkDegradesOnUnreachable(t *testing.T) {
	noAmbientProxy(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// A closed port: nothing is listening.
	got, ok := resolveTikTokShortLink(ctx, "http://127.0.0.1:1/short", "")
	if ok {
		t.Error("expected ok=false when the host is unreachable")
	}
	if got != "http://127.0.0.1:1/short" {
		t.Errorf("should return the original URL, got %q", got)
	}
}

// TestResolveShortLinkBoundedRedirects ensures a redirect loop terminates rather
// than hanging until the outer deadline.
func TestResolveShortLinkBoundedRedirects(t *testing.T) {
	noAmbientProxy(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/loop", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loop", http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	// Must return, not loop forever.
	_, ok := resolveTikTokShortLink(ctx, srv.URL+"/loop", "")
	if elapsed := time.Since(start); elapsed > 8*time.Second {
		t.Errorf("redirect loop took %v; the hop bound did not stop it", elapsed)
	}
	if ok {
		t.Error("a redirect loop identifies no post, so ok must be false")
	}
}

// TestResolveShortLinkRejectsNonPostDestination covers the case the previous
// return-anything-as-resolved behaviour got wrong: a proxy error page, a captcha
// wall or a redirect to the site homepage all produce a response but identify no
// post. Claiming "resolved" for those would route a photo post down the video
// path with a URL that is not a post at all.
func TestResolveShortLinkRejectsNonPostDestination(t *testing.T) {
	noAmbientProxy(t)

	cases := map[string]struct {
		status int
		body   string
	}{
		"proxy error page": {http.StatusBadGateway, "502 Bad Gateway"},
		"captcha wall":     {http.StatusOK, "<html>are you a robot?</html>"},
		"site homepage":    {http.StatusOK, "<html>welcome to tiktok</html>"},
		"empty response":   {http.StatusOK, ""},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/short", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.status)
				_, _ = w.Write([]byte(c.body))
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			got, ok := resolveTikTokShortLink(ctx, srv.URL+"/short", "")
			if ok {
				t.Errorf("%s: ok must be false when no post URL was identified", name)
			}
			if got != srv.URL+"/short" {
				t.Errorf("%s: should return the original URL, got %q", name, got)
			}
		})
	}
}

// TestResolveShortLinkRejectsRedirectToNonPost is the redirect variant: the chain
// ends on tiktok.com, but not at a post, so nothing was resolved.
func TestResolveShortLinkRejectsRedirectToNonPost(t *testing.T) {
	noAmbientProxy(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/short", func(w http.ResponseWriter, r *http.Request) {
		// Same host, but a path that is not a post URL.
		http.Redirect(w, r, "/discover", http.StatusFound)
	})
	mux.HandleFunc("/discover", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html>trending</html>"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	got, ok := resolveTikTokShortLink(ctx, srv.URL+"/short", "")
	if ok {
		t.Error("a redirect to a non-post path identifies no post; ok must be false")
	}
	if got != srv.URL+"/short" {
		t.Errorf("should return the original URL, got %q", got)
	}
}

// TestResolveShortLinkHonorsContextDeadline verifies a caller deadline aborts the
// resolution instead of waiting for the internal timeout.
func TestResolveShortLinkHonorsContextDeadline(t *testing.T) {
	noAmbientProxy(t)
	// 10.255.255.1 is unroutable, so the dial would otherwise hang.
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	got, ok := resolveTikTokShortLink(ctx, "http://10.255.255.1:9/short", "")
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("ignored the caller deadline, took %v", elapsed)
	}
	if ok {
		t.Error("expected ok=false for an unreachable host")
	}
	if got != "http://10.255.255.1:9/short" {
		t.Errorf("should return the original URL, got %q", got)
	}
}

// TestResolveShortLinkStopsAtPostURL verifies resolution halts the moment a hop
// identifies a post, rather than continuing to follow redirects.
//
// This matters for cost: the caller fetches the page itself when scraping a photo
// post, so following the final hop would download a ~400 KB TikTok page twice on
// every photo link. It also bounds exposure to an external host during routing.
//
// The chain is served locally and counted, with the final hop pointing at a real
// TikTok URL. Stopping at that hop means no outbound request is ever made, so a
// fast return with the post URL proves the boundary held.
func TestResolveShortLinkStopsAtPostURL(t *testing.T) {
	noAmbientProxy(t)
	const post = "https://www.tiktok.com/@lil.mill_000/photo/7633745141773257991"
	var hops int32

	mux := http.NewServeMux()
	mux.HandleFunc("/short", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hops, 1)
		http.Redirect(w, r, "/middle", http.StatusFound)
	})
	mux.HandleFunc("/middle", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hops, 1)
		http.Redirect(w, r, post, http.StatusFound)
	})
	// Should never be reached: resolution must stop at the post URL above.
	mux.HandleFunc("/must-not-be-fetched", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hops, 1)
		_, _ = w.Write(make([]byte, 512<<10))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	got, ok := resolveTikTokShortLink(ctx, srv.URL+"/short", "")
	elapsed := time.Since(start)

	if !ok {
		t.Fatal("resolution reported failure")
	}
	if got != post {
		t.Errorf("resolved = %q, want %q", got, post)
	}
	// Two local hops are followed; the third would be the outbound TikTok page.
	if n := atomic.LoadInt32(&hops); n != 2 {
		t.Errorf("local hops = %d, want 2 (must not continue past the post URL)", n)
	}
	// Reaching tiktok.com would either fail in this sandbox or take measurably
	// longer than two loopback redirects.
	if elapsed > 5*time.Second {
		t.Errorf("took %v, suggesting an outbound fetch past the stop point", elapsed)
	}
}

func TestResolveShortLinkInvalidProxyDegrades(t *testing.T) {
	noAmbientProxy(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	got, ok := resolveTikTokShortLink(ctx, "https://vt.tiktok.com/abc", "not a proxy url")
	if ok {
		t.Error("expected ok=false with an unusable proxy specification")
	}
	if got != "https://vt.tiktok.com/abc" {
		t.Errorf("should return the original URL, got %q", got)
	}
}
