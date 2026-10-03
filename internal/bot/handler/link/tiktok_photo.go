package link

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	proxypkg "famoria/internal/pkg/proxy"
)

// maxPhotoPostImages caps how many images of a TikTok photo post are sent.
// Telegram albums accept up to 10, but a photo post can hold dozens of images
// and re-uploading all of them is slow and floods the chat.
const maxPhotoPostImages = 5

const (
	// scrapeTimeout bounds the page fetch; photo pages are ~400 KB.
	scrapeTimeout = 30 * time.Second
	// resolveTimeout bounds following a share link's redirects to discover the
	// canonical post URL.
	resolveTimeout = 20 * time.Second
	// imageTimeout bounds a single image download.
	imageTimeout = 30 * time.Second
	// maxImageBytes rejects anything absurd, guarding against a changed page
	// structure handing us an HTML error page instead of a JPEG.
	maxImageBytes = 10 << 20
)

// browserUA is required: TikTok serves the rehydration JSON to real browsers
// and a stripped-down or bot user agent gets a page without the post data.
const browserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/120.0 Safari/537.36"

// rehydrationScript matches the embedded state blob TikTok ships on every page.
var rehydrationScript = regexp.MustCompile(
	`<script id="__UNIVERSAL_DATA_FOR_REHYDRATION__" type="application/json">(.*?)</script>`,
)

// photoPost is the result of scraping a TikTok photo (slideshow) post.
type photoPost struct {
	Title string
	// Images are local file paths, in album order, already limited to
	// maxPhotoPostImages.
	Images []string
}

// isTikTokPhotoPost reports whether the URL addresses a TikTok photo post
// rather than a video. yt-dlp has no extractor for these — its TikTok matcher
// only accepts /video/, /embed/ and /share/video/ paths — so they are handled
// by scraping the page instead.
//
// Note this only classifies URLs whose path is already known. A share link like
// vt.tiktok.com/ZSbmeYhHC carries no /photo/ segment even when it points at a
// photo post, so callers must resolve the link first via
// resolveTikTokShortLink; classifying the short form sends photo posts down the
// video path, where yt-dlp rejects them as "Unsupported URL".
func isTikTokPhotoPost(rawURL string) bool {
	if !isTikTok(rawURL) {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(u.Path), "/photo/")
}

// tiktokShortLinkHosts are the link-shortener hosts TikTok uses for shares.
var tiktokShortLinkHosts = []string{"vt.tiktok.com", "vm.tiktok.com"}

// maxShortLinkRedirects bounds the redirect walk when resolving a share link.
const maxShortLinkRedirects = 5

// isTikTokShortLink reports whether the URL is a share link that must be resolved
// before its post type is known.
func isTikTokShortLink(rawURL string) bool {
	host, ok := urlHost(rawURL)
	if !ok {
		return false
	}
	for _, h := range tiktokShortLinkHosts {
		if host == h {
			return true
		}
	}
	return false
}

// isTikTokPostURL reports whether a URL already identifies a specific post, which
// is the signal to stop following redirects.
func isTikTokPostURL(rawURL string) bool {
	if !isTikTok(rawURL) {
		return false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	path := strings.ToLower(u.Path)
	return strings.Contains(path, "/photo/") || strings.Contains(path, "/video/")
}

// resolveTikTokShortLink follows a share link's redirects and returns the
// canonical post URL.
//
// It stops at the first URL that identifies a post and returns
// http.ErrUseLastResponse at that point, so the page body is never downloaded —
// the caller is about to fetch that same page anyway when scraping a photo post.
//
// On any failure the original URL is returned unchanged rather than an error: a
// short link is still usable by yt-dlp for video posts, so failing to classify it
// should degrade to the existing behaviour, not break the download.
func resolveTikTokShortLink(ctx context.Context, rawURL, proxySpec string) (string, bool) {
	client, err := proxyClient(proxySpec)
	if err != nil {
		return rawURL, false
	}

	resolveCtx, cancel := context.WithTimeout(ctx, resolveTimeout)
	defer cancel()

	// Copy the client so the shared one keeps its default redirect behaviour.
	resolving := *client
	var last string
	resolving.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		last = req.URL.String()
		// Stop as soon as the post is identified, or the chain looks endless.
		if isTikTokPostURL(last) || len(via) >= maxShortLinkRedirects {
			return http.ErrUseLastResponse
		}
		return nil
	}

	req, err := http.NewRequestWithContext(resolveCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		return rawURL, false
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := resolving.Do(req)
	if err != nil {
		return rawURL, false
	}
	defer resp.Body.Close()
	// Drain a little so the connection can be reused, without reading a full page.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	// Only report success when a post URL was actually identified. A proxy error
	// page, a captcha wall or a redirect to tiktok.com's homepage all produce a
	// response but no post, and claiming "resolved" for those would send routing
	// down the video path with a URL that is not a post at all.
	if last != "" && isTikTokPostURL(last) {
		return last, true
	}
	if resp.Request != nil && resp.Request.URL != nil {
		if final := resp.Request.URL.String(); isTikTokPostURL(final) {
			return final, true
		}
	}
	return rawURL, false
}

// photoPathSegment matches the /photo/ segment of a post path, case
// insensitively. Only that segment is rewritten: lowercasing the whole path
// would break usernames that contain capitals, since TikTok paths are
// case-sensitive.
var photoPathSegment = regexp.MustCompile(`(?i)/photo/`)

// videoURLFromPhoto rewrites a /photo/ URL to its /video/ equivalent.
//
// This is not a mistake: TikTok serves the same rehydration payload on both
// paths, and only the /video/ form includes the itemStruct that carries the
// imagePost list. The /photo/ page renders without it.
func videoURLFromPhoto(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.Path = photoPathSegment.ReplaceAllString(u.Path, "/video/")
	// Query parameters are share attribution and can defeat the lookup.
	u.RawQuery = ""
	return u.String()
}

// scrapeTikTokPhotoPost fetches a TikTok photo post and downloads its images
// into dir.
//
// It returns an error when the page carries no image list, which also covers
// the case of a private or removed post.
func scrapeTikTokPhotoPost(ctx context.Context, rawURL, dir string, proxySpec string) (*photoPost, error) {
	// Resolve the HTTP client once, before any request. Doing it per image instead
	// would let a malformed proxy specification surface as "no images downloaded"
	// — the loop below skips individual failures — hiding the real cause.
	client, err := proxyClient(proxySpec)
	if err != nil {
		return nil, fmt.Errorf("tiktok photo post: %w", err)
	}

	page, err := fetchPage(ctx, videoURLFromPhoto(rawURL), client)
	if err != nil {
		return nil, err
	}

	images, title, err := parsePhotoPost(page)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("tiktok photo post has no images")
	}
	if len(images) > maxPhotoPostImages {
		images = images[:maxPhotoPostImages]
	}

	post := &photoPost{Title: title}
	for i, imgURL := range images {
		path := filepath.Join(dir, fmt.Sprintf("photo_%02d.jpg", i+1))
		if err := downloadImage(ctx, imgURL, path, client); err != nil {
			// A single failed image should not abort the album; skip it. If every
			// image fails the caller sees an empty list and falls through.
			continue
		}
		post.Images = append(post.Images, path)
	}
	if len(post.Images) == 0 {
		return nil, fmt.Errorf("tiktok photo post: no images downloaded")
	}
	return post, nil
}

// fetchPage downloads a TikTok page and returns its body.
func fetchPage(ctx context.Context, pageURL string, client *http.Client) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, scrapeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tiktok page returned %s", resp.Status)
	}

	var body io.Reader = resp.Body
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer gz.Close()
		body = gz
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// parsePhotoPost pulls the image URLs and the post description out of the
// rehydration JSON.
//
// The path is __DEFAULT_SCOPE__ → webapp.video-detail → itemInfo → itemStruct,
// where imagePost.images[].imageURL.urlList[] holds the CDN links. Each image
// lists several mirrors; the first reachable one is used.
func parsePhotoPost(page string) ([]string, string, error) {
	m := rehydrationScript.FindStringSubmatch(page)
	if m == nil {
		return nil, "", fmt.Errorf("rehydration script not found in page")
	}

	var root struct {
		DefaultScope map[string]any `json:"__DEFAULT_SCOPE__"`
	}
	if err := json.Unmarshal([]byte(m[1]), &root); err != nil {
		return nil, "", fmt.Errorf("rehydration json: %w", err)
	}

	itemStruct, err := digItemStruct(root.DefaultScope)
	if err != nil {
		return nil, "", err
	}

	title, _ := itemStruct["desc"].(string)

	// Photo posts only: a video post has no imagePost key.
	imagePost, ok := itemStruct["imagePost"].(map[string]any)
	if !ok {
		return nil, "", fmt.Errorf("post has no imagePost (not a photo post?)")
	}
	images, ok := imagePost["images"].([]any)
	if !ok || len(images) == 0 {
		return nil, "", fmt.Errorf("imagePost has no images")
	}

	var urls []string
	for _, raw := range images {
		img, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if u := firstImageURL(img); u != "" {
			urls = append(urls, u)
		}
	}
	return urls, title, nil
}

// digItemStruct navigates the rehydration scope down to the post's itemStruct.
//
// Errors include TikTok's own status code and message when present. That detail
// is what distinguishes the cases that look identical from the outside: a removed
// or private post, a regional block, and a consent/anti-bot stub page all arrive
// with webapp.video-detail present but no itemInfo, and "itemInfo missing" alone
// gives nothing to act on.
func digItemStruct(scope map[string]any) (map[string]any, error) {
	detail, ok := scope["webapp.video-detail"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("webapp.video-detail missing (page did not contain the post; "+
			"scopes present: %s)", strings.Join(sortedKeys(scope), ", "))
	}

	info, ok := detail["itemInfo"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("itemInfo missing (%s)", describeTikTokStatus(detail))
	}

	itemStruct, ok := info["itemStruct"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("itemStruct missing (%s)", describeTikTokStatus(detail))
	}
	return itemStruct, nil
}

// describeTikTokStatus renders the status fields TikTok returns alongside a
// failed lookup, or says so when there are none.
func describeTikTokStatus(detail map[string]any) string {
	var parts []string
	if code, ok := detail["statusCode"]; ok {
		parts = append(parts, fmt.Sprintf("statusCode=%v", code))
	}
	if msg, ok := detail["statusMsg"].(string); ok && msg != "" {
		parts = append(parts, fmt.Sprintf("statusMsg=%q", msg))
	}
	if len(parts) == 0 {
		return "no status fields; scopes: " + strings.Join(sortedKeys(detail), ", ")
	}
	return strings.Join(parts, ", ")
}

// sortedKeys returns a map's keys in a stable order, so error text is comparable
// between runs instead of being reshuffled by Go's map iteration.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// firstImageURL returns the first usable CDN URL for one image entry.
func firstImageURL(img map[string]any) string {
	imageURL, ok := img["imageURL"].(map[string]any)
	if !ok {
		return ""
	}
	list, ok := imageURL["urlList"].([]any)
	if !ok {
		return ""
	}
	for _, v := range list {
		if s, ok := v.(string); ok && strings.HasPrefix(s, "http") {
			return s
		}
	}
	return ""
}

// downloadImage fetches one CDN image to path, rejecting responses that are not
// images.
func downloadImage(ctx context.Context, imgURL, path string, client *http.Client) error {
	ctx, cancel := context.WithTimeout(ctx, imageTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Referer", "https://www.tiktok.com/")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("image returned %s", resp.Status)
	}

	ctype := resp.Header.Get("Content-Type")
	if ctype != "" && !strings.HasPrefix(strings.ToLower(ctype), "image/") {
		return fmt.Errorf("unexpected content type %q", ctype)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := io.Copy(f, io.LimitReader(resp.Body, maxImageBytes))
	if err != nil {
		_ = os.Remove(path)
		return err
	}
	if n == 0 {
		_ = os.Remove(path)
		return fmt.Errorf("empty image body")
	}
	return nil
}

// proxyClient builds an HTTP client for the scraper, optionally routed through
// the configured proxy.
//
// It parses via proxy.Parse rather than url.Parse directly: the documented
// configuration form "login:pass@host:port" carries no scheme, and url.Parse
// reads "login" as the scheme and leaves Host empty, which would silently
// produce a client that never uses the proxy.
//
// A malformed specification is an error rather than a fallback to a direct
// connection. Going direct would defeat the purpose of the proxy on a network
// where the destination is blocked, and would look like a TikTok failure rather
// than a misconfiguration.
//
// With an empty specification the default client is returned, which still honours
// the ambient HTTP_PROXY/HTTPS_PROXY environment variables. That is deliberate:
// operators who proxy the whole container by environment keep working without
// also setting PROXY_URL. It does mean "no proxy configured" is not the same as
// "guaranteed direct" — callers must not rely on the latter.
func proxyClient(proxySpec string) (*http.Client, error) {
	if proxySpec == "" {
		return http.DefaultClient, nil
	}
	ep, err := proxypkg.Parse(proxySpec)
	if err != nil {
		return nil, err
	}
	return ep.HTTPClient(), nil
}
