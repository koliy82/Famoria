package link

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lrstanley/go-ytdlp"
	"github.com/mymmrac/telego"
)

// These tests hit the network and are skipped unless LINK_LIVE_TESTS=1 is set,
// so normal `go test ./...` runs stay fast and hermetic.

func liveEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("LINK_LIVE_TESTS") != "1" {
		t.Skip("set LINK_LIVE_TESTS=1 to run network tests")
	}
}

func testConfig() ytConfig {
	cfg := ytConfig{cookiesFile: "cookies/cookies.txt"}
	if _, err := os.Stat(cfg.cookiesFile); err != nil {
		cfg.cookiesFile = ""
	}
	return cfg
}

// TestVideoStrategiesOrder asserts the platform-specific ordering: YouTube must
// try DASH first (its combined stream is capped at 360p), while TikTok must try
// the combined stream first (merging separate streams there drops the audio).
func TestVideoStrategiesOrder(t *testing.T) {
	yt := videoStrategies(true)
	if yt[0].name != "dash-h264" {
		t.Errorf("youtube first strategy = %q, want dash-h264", yt[0].name)
	}

	tk := videoStrategies(false)
	if tk[0].name != "combined-h264" {
		t.Errorf("tiktok first strategy = %q, want combined-h264", tk[0].name)
	}

	for _, set := range [][]videoStrategy{yt, tk} {
		if len(set) != 3 {
			t.Fatalf("expected 3 strategies, got %d", len(set))
		}
		if set[len(set)-1].name != "best" {
			t.Errorf("last strategy = %q, want best fallback", set[len(set)-1].name)
		}
	}
}

// TestTelegramFormatSort verifies the resolution cap is only applied when a
// height is requested, so the first attempt stays uncapped and full quality.
func TestTelegramFormatSort(t *testing.T) {
	if got := telegramFormatSort(0); !strings.HasPrefix(got, "res,") {
		t.Errorf("height 0 should not cap resolution, got %q", got)
	}
	if got := telegramFormatSort(720); !strings.HasPrefix(got, "res:720,") {
		t.Errorf("height 720 should cap to res:720, got %q", got)
	}

	// Both codec namings must be preferred. TikTok reports "h264"/"aac" while
	// YouTube reports "avc1.42001E"/"mp4a.40.2"; naming only one family silently
	// loses the other platform's formats.
	for _, s := range []string{telegramFormatSort(0), telegramFormatSort(480)} {
		for _, want := range []string{"avc1", "h264", "mp4a", "aac"} {
			if !strings.Contains(s, want) {
				t.Errorf("sort must prefer %q for Telegram, got %q", want, s)
			}
		}
	}
}

// TestCombinedSelectorMatchesBothCodecNamings is the regression guard for the bug
// where a selector written only as [vcodec^=h264][acodec^=aac] matched nothing on
// YouTube, making yt-dlp answer "Requested format is not available" and the whole
// cascade fail.
func TestCombinedSelectorMatchesBothCodecNamings(t *testing.T) {
	s := combinedSelector
	for _, want := range []string{"avc1", "mp4a", "h264", "aac"} {
		if !strings.Contains(s, want) {
			t.Errorf("combinedSelector is missing %q: %q", want, s)
		}
	}
	// avc1/mp4a must be tried first, since that is the YouTube naming and the
	// h264/aac branch would otherwise short-circuit on TikTok-style catalogs.
	if strings.Index(s, "avc1") > strings.Index(s, "h264") {
		t.Errorf("avc1 should be preferred before h264: %q", s)
	}
	// A trailing /best keeps a usable fallback for catalogs that name codecs
	// differently again, so the selector never resolves to "not available".
	if !strings.HasSuffix(s, "/best") {
		t.Errorf("combinedSelector should end with a /best fallback: %q", s)
	}
}

// TestVideoStrategiesAlwaysHaveCombinedFirst checks that whichever platform is
// being served, a progressive (video+audio) strategy is reachable — the case that
// still works when DASH downloads are refused with 403.
func TestVideoStrategiesAlwaysHaveCombinedFirst(t *testing.T) {
	for _, isYT := range []bool{true, false} {
		strats := videoStrategies(isYT)
		found := false
		for _, s := range strats {
			if s.name == "combined-h264" {
				found = true
			}
		}
		if !found {
			t.Errorf("isYT=%v: no combined-h264 strategy, so a 403 on DASH leaves nothing sendable", isYT)
		}
	}

	// TikTok must try the combined stream first: merging separate streams there
	// yields a file with no audio track at all.
	tk := videoStrategies(false)
	if tk[0].name != "combined-h264" {
		t.Errorf("tiktok first strategy = %q, want combined-h264", tk[0].name)
	}
}

// TestResolutionLadderStartsUncapped guards the quality regression: the ladder
// must begin at 0 (uncapped) so videos are never downscaled preemptively.
func TestResolutionLadderStartsUncapped(t *testing.T) {
	if len(resolutionLadder) == 0 || resolutionLadder[0] != 0 {
		t.Errorf("ladder must start uncapped, got %v", resolutionLadder)
	}
}

// TestNoPlayerClientOverride pins the quality fix: applying params must not set
// a player_client, which collapses YouTube's format list to 360p.
func TestNoPlayerClientOverride(t *testing.T) {
	cmd := applyParams(ytdlp.New(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ", ytConfig{})
	flags := cmd.BuildCommand(context.Background(), "x").Args
	for _, a := range flags {
		if strings.Contains(a, "player_client") {
			t.Errorf("player_client must not be pinned, found %q", a)
		}
	}
}

func TestApplyParamsCookiesOnlyForYouTube(t *testing.T) {
	cfg := ytConfig{cookiesFile: "/tmp/cookies.txt", proxy: "http://p:1", proxyMode: proxyModeYouTube}

	yt := cmdFlags(applyParams(ytdlp.New(), "https://youtu.be/abc123456", cfg))
	if !strings.Contains(yt, "/tmp/cookies.txt") {
		t.Error("youtube should get cookies")
	}
	if !strings.Contains(yt, "http://p:1") {
		t.Error("youtube-mode proxy should apply to youtube URLs")
	}

	tk := cmdFlags(applyParams(ytdlp.New(), "https://www.tiktok.com/@u/video/1", cfg))
	if strings.Contains(tk, "/tmp/cookies.txt") {
		t.Error("tiktok should not get youtube cookies")
	}
	if strings.Contains(tk, "http://p:1") {
		t.Error("youtube-only proxy must not apply to tiktok")
	}
}

func TestApplyParamsProxyModeAll(t *testing.T) {
	cfg := ytConfig{proxy: "http://p:2", proxyMode: proxyModeAll}
	tk := cmdFlags(applyParams(ytdlp.New(), "https://www.tiktok.com/@u/video/1", cfg))
	if !strings.Contains(tk, "http://p:2") {
		t.Error("mode=true should proxy every URL")
	}

	cfg.proxyMode = proxyModeFalse
	if strings.Contains(cmdFlags(applyParams(ytdlp.New(), "https://youtu.be/abc123456", cfg)), "http://p:2") {
		t.Error("mode=false must never proxy")
	}
}

func TestHostDetection(t *testing.T) {
	ytURLs := []string{
		"https://www.youtube.com/watch?v=abc", "https://youtu.be/abc",
		"https://m.youtube.com/x", "https://music.youtube.com/x",
		"https://www.youtube.com/shorts/abc",
	}
	for _, u := range ytURLs {
		if !isYouTube(u) {
			t.Errorf("isYouTube(%q) = false", u)
		}
		if isTikTok(u) {
			t.Errorf("isTikTok(%q) should be false for a youtube URL", u)
		}
	}

	tkURLs := []string{
		"https://www.tiktok.com/@u/video/1", "https://vt.tiktok.com/ABC/",
		"https://vm.tiktok.com/ABC",
	}
	for _, u := range tkURLs {
		if !isTikTok(u) {
			t.Errorf("isTikTok(%q) = false", u)
		}
		if isYouTube(u) {
			t.Errorf("isYouTube(%q) should be false for tiktok", u)
		}
	}

	for _, u := range []string{"", "not a url", "https://vimeo.com/1"} {
		if isYouTube(u) || isTikTok(u) {
			t.Errorf("unexpected host match for %q", u)
		}
	}
}

func TestVideoURLFromPhoto(t *testing.T) {
	cases := map[string]string{
		"https://www.tiktok.com/@Lil.Mill_000/photo/7633?_r=1&_t=X": "https://www.tiktok.com/@Lil.Mill_000/video/7633",
		"https://www.tiktok.com/@user/video/7633":                   "https://www.tiktok.com/@user/video/7633",
	}
	for in, want := range cases {
		if got := videoURLFromPhoto(in); got != want {
			t.Errorf("videoURLFromPhoto(%q) = %q, want %q", in, got, want)
		}
	}
	// Case must be preserved: TikTok usernames with capitals are valid.
	got := videoURLFromPhoto("https://www.tiktok.com/@UserName/photo/123")
	if !strings.Contains(got, "@UserName") {
		t.Errorf("username case was mangled: %q", got)
	}
}

func TestIsTikTokPhotoPost(t *testing.T) {
	if !isTikTokPhotoPost("https://www.tiktok.com/@u/photo/123?_r=1") {
		t.Error("photo post not detected")
	}
	if isTikTokPhotoPost("https://www.tiktok.com/@u/video/123") {
		t.Error("video post misdetected as photo")
	}
	if isTikTokPhotoPost("https://youtu.be/abc123456") {
		t.Error("youtube misdetected as tiktok photo")
	}
}

func TestFindMP4PicksLargest(t *testing.T) {
	dir := t.TempDir()
	mustWrite := func(name string, size int) {
		if err := os.WriteFile(filepath.Join(dir, name), make([]byte, size), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("video.f1.mp4", 100)
	mustWrite("video.mp4", 5000)
	mustWrite("notes.txt", 9000)

	got, err := findMP4(dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "video.mp4" {
		t.Errorf("findMP4 = %q, want the largest mp4", filepath.Base(got))
	}

	if _, err := findMP4(t.TempDir()); err == nil {
		t.Error("findMP4 on an empty dir should error")
	}
}

// TestClearDirPreventsStalePick pins the bug where a rejected over-50MB file
// from an earlier attempt stayed in the temp dir and then won findMP4's
// largest-file selection, sending a file Telegram refuses.
func TestClearDirPreventsStalePick(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "video.mp4")
	if err := os.WriteFile(stale, make([]byte, 60_000_000), 0o644); err != nil {
		t.Fatal(err)
	}

	clearDir(dir)

	// The directory itself must survive, so it stays usable for the next attempt.
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("clearDir removed the directory itself: %v", err)
	}
	if _, err := findMP4(dir); err == nil {
		t.Error("stale mp4 survived clearDir and would be picked again")
	}
}

func TestClearDirOnMissingDirIsHarmless(t *testing.T) {
	clearDir(filepath.Join(t.TempDir(), "does-not-exist"))
}

func TestParsePhotoPostRejectsGarbage(t *testing.T) {
	if _, _, err := parsePhotoPost("<html>no state here</html>"); err == nil {
		t.Error("expected an error for a page without the rehydration script")
	}

	page := `<script id="__UNIVERSAL_DATA_FOR_REHYDRATION__" type="application/json">{"__DEFAULT_SCOPE__":{}}</script>`
	if _, _, err := parsePhotoPost(page); err == nil {
		t.Error("expected an error when webapp.video-detail is missing")
	}

	// A video post has itemStruct but no imagePost key.
	video := `<script id="__UNIVERSAL_DATA_FOR_REHYDRATION__" type="application/json">` +
		`{"__DEFAULT_SCOPE__":{"webapp.video-detail":{"itemInfo":{"itemStruct":{"desc":"hi","video":{}}}}}}</script>`
	if _, _, err := parsePhotoPost(video); err == nil {
		t.Error("expected an error for a video post without imagePost")
	}
}

// TestParsePhotoPostErrorCarriesTikTokStatus pins the diagnostics improvement.
//
// A removed post, a private post, a regional block and an anti-bot stub all look
// identical from here: webapp.video-detail is present but itemInfo is absent.
// Surfacing TikTok's own statusCode/statusMsg is what makes those separable in
// the logs instead of every one reading "itemInfo missing".
func TestParsePhotoPostErrorCarriesTikTokStatus(t *testing.T) {
	page := func(inner string) string {
		return `<script id="__UNIVERSAL_DATA_FOR_REHYDRATION__" type="application/json">` +
			`{"__DEFAULT_SCOPE__":{"webapp.video-detail":` + inner + `}}</script>`
	}

	cases := []struct {
		name       string
		detail     string
		wantInErr  []string
		dontWantIn []string
	}{
		{
			name:      "item removed",
			detail:    `{"statusCode":10204,"statusMsg":"Video is unavailable"}`,
			wantInErr: []string{"10204", "Video is unavailable"},
		},
		{
			name:      "blocked or private",
			detail:    `{"statusCode":10221,"statusMsg":"Video is private"}`,
			wantInErr: []string{"10221", "private"},
		},
		{
			name:       "no status fields at all",
			detail:     `{"somethingElse":true}`,
			wantInErr:  []string{"itemInfo missing", "somethingElse"},
			dontWantIn: []string{"statusCode="},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := parsePhotoPost(page(c.detail))
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, want := range c.wantInErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q should mention %q", err, want)
				}
			}
			for _, unwanted := range c.dontWantIn {
				if strings.Contains(err.Error(), unwanted) {
					t.Errorf("error %q should not claim %q", err, unwanted)
				}
			}
		})
	}
}

// TestSortedKeysIsStable keeps error text comparable between runs: Go randomizes
// map iteration, so an unsorted key list would produce a different message on
// every failure.
func TestSortedKeysIsStable(t *testing.T) {
	m := map[string]any{"zebra": 1, "alpha": 2, "mid": 3}
	want := "alpha, mid, zebra"
	for i := 0; i < 20; i++ {
		if got := strings.Join(sortedKeys(m), ", "); got != want {
			t.Fatalf("sortedKeys = %q, want %q", got, want)
		}
	}
	if len(sortedKeys(map[string]any{})) != 0 {
		t.Error("empty map should yield no keys")
	}
}

func TestParsePhotoPostExtractsImages(t *testing.T) {
	page := `<script id="__UNIVERSAL_DATA_FOR_REHYDRATION__" type="application/json">` +
		`{"__DEFAULT_SCOPE__":{"webapp.video-detail":{"itemInfo":{"itemStruct":{` +
		`"desc":"my title",` +
		`"imagePost":{"images":[` +
		`{"imageURL":{"urlList":["https://cdn/a.jpeg","https://cdn/mirror/a.jpeg"]}},` +
		`{"imageURL":{"urlList":["https://cdn/b.jpeg"]}}` +
		`]}}}}}}</script>`

	urls, title, err := parsePhotoPost(page)
	if err != nil {
		t.Fatal(err)
	}
	if title != "my title" {
		t.Errorf("title = %q, want %q", title, "my title")
	}
	if len(urls) != 2 {
		t.Fatalf("urls = %d, want 2", len(urls))
	}
	// The first mirror is preferred.
	if urls[0] != "https://cdn/a.jpeg" {
		t.Errorf("urls[0] = %q, want the first mirror", urls[0])
	}
}

func TestFirstImageURLSkipsBadEntries(t *testing.T) {
	if got := firstImageURL(map[string]any{}); got != "" {
		t.Errorf("empty entry should yield no URL, got %q", got)
	}
	if got := firstImageURL(map[string]any{"imageURL": map[string]any{"urlList": []any{"ftp://x"}}}); got != "" {
		t.Errorf("non-http URL should be skipped, got %q", got)
	}
}

func TestBuildCaption(t *testing.T) {
	got := buildCaption("ru", "Название", "https://youtu.be/x")
	if !strings.Contains(got, "Название") {
		t.Errorf("caption missing title: %q", got)
	}
	if !strings.Contains(got, "converted by Famoria") {
		t.Errorf("caption missing attribution: %q", got)
	}
	if !strings.Contains(got, `href="https://youtu.be/x"`) {
		t.Errorf("caption missing source link: %q", got)
	}

	// No title falls back to the bare form.
	if got := buildCaption("ru", "", "https://youtu.be/x"); strings.Contains(got, "<b></b>") {
		t.Errorf("empty title should not render an empty bold tag: %q", got)
	}
}

// TestBuildCaptionTruncatesLongTitle guards Telegram's 1024 char caption limit:
// an over-long caption makes the whole send fail with 400.
func TestBuildCaptionTruncatesLongTitle(t *testing.T) {
	long := strings.Repeat("слово ", 400)
	got := buildCaption("ru", long, "https://youtu.be/x")
	if len(got) > captionLimit {
		t.Errorf("caption is %d bytes, exceeds the %d limit", len(got), captionLimit)
	}
	if !strings.Contains(got, "converted by Famoria") {
		t.Error("attribution must survive truncation")
	}
	if !strings.Contains(got, "href=") {
		t.Error("source link must survive truncation")
	}

	// HTML-escaping must not let a crafted title break out of the caption.
	if got := buildCaption("ru", `<script>alert("x")</script>`, "https://youtu.be/x"); strings.Contains(got, "<script>") {
		t.Error("title must be HTML-escaped")
	}
}

func TestLargestPhotoFileID(t *testing.T) {
	// Empty message yields no ID rather than panicking.
	if got := largestPhotoFileID(telego.Message{}); got != "" {
		t.Errorf("expected empty for a message with no photos, got %q", got)
	}
	// Telegram lists sizes ascending, so the last entry is the largest.
	m := telego.Message{Photo: []telego.PhotoSize{
		{FileID: "small"}, {FileID: "medium"}, {FileID: "large"},
	}}
	if got := largestPhotoFileID(m); got != "large" {
		t.Errorf("largestPhotoFileID = %q, want large", got)
	}
}

func cmdFlags(cmd *ytdlp.Command) string {
	return strings.Join(cmd.BuildCommand(context.Background(), "https://x").Args, " ")
}

// fmt helpers for building synthetic format catalogs.

func f64(v float64) *float64 { return &v }
func iPtr(v int) *int        { return &v }
func sPtr(v string) *string  { return &v }

func TestSelectFormatSpecPrefersHighestH264UnderBudget(t *testing.T) {
	dur := 60.0
	info := &ytdlp.ExtractedInfo{
		Duration: &dur,
		Formats: []*ytdlp.ExtractedFormat{
			// Progressive h264+aac at 360p — valid but not the best picture.
			{FormatID: sPtr("18"), VCodec: sPtr("avc1.42"), ACodec: sPtr("mp4a.40.2"),
				Height: f64(360), FileSize: iPtr(5_000_000)},
			// Video-only h264 at 1080p plus AAC audio: the winner.
			{FormatID: sPtr("137"), VCodec: sPtr("avc1.640028"), Height: f64(1080),
				FileSize: iPtr(20_000_000)},
			{FormatID: sPtr("140"), ACodec: sPtr("mp4a.40.2"), FileSize: iPtr(1_000_000)},
			// HEVC at 1080p must be rejected: Telegram cannot play it.
			{FormatID: sPtr("248"), VCodec: sPtr("vp9"), Height: f64(1080), FileSize: iPtr(15_000_000)},
			// Opus audio must not be paired: it plays silent in mp4.
			{FormatID: sPtr("251"), ACodec: sPtr("opus"), FileSize: iPtr(900_000)},
		},
	}

	got := selectFormatSpec(info)
	if got != "137+140" {
		t.Errorf("selectFormatSpec = %q, want %q", got, "137+140")
	}
}

// TestSelectFormatSpecRespectsBudget guards the 50 MB upload limit: a huge
// format must be skipped in favour of one that fits, otherwise Telegram rejects
// the send and the whole job fails.
func TestSelectFormatSpecRespectsBudget(t *testing.T) {
	dur := 1800.0
	info := &ytdlp.ExtractedInfo{
		Duration: &dur,
		Formats: []*ytdlp.ExtractedFormat{
			{FormatID: sPtr("big"), VCodec: sPtr("avc1"), ACodec: sPtr("mp4a"),
				Height: f64(1080), FileSize: iPtr(400_000_000)},
			{FormatID: sPtr("fits"), VCodec: sPtr("avc1"), ACodec: sPtr("mp4a"),
				Height: f64(480), FileSize: iPtr(30_000_000)},
		},
	}
	if got := selectFormatSpec(info); got != "fits" {
		t.Errorf("selectFormatSpec = %q, want %q", got, "fits")
	}
}

// TestSelectFormatSpecEmptyWhenNoSizes verifies the fallback contract: with no
// size information the caller must use the strategy cascade instead.
func TestSelectFormatSpecEmptyWhenNoSizes(t *testing.T) {
	info := &ytdlp.ExtractedInfo{
		Formats: []*ytdlp.ExtractedFormat{
			{FormatID: sPtr("18"), VCodec: sPtr("avc1"), ACodec: sPtr("mp4a"), Height: f64(360)},
		},
	}
	if got := selectFormatSpec(info); got != "" {
		t.Errorf("expected empty spec when sizes are unknown, got %q", got)
	}
	if got := selectFormatSpec(nil); got != "" {
		t.Errorf("expected empty spec for nil info, got %q", got)
	}
}

// TestSelectFormatSpecUsesBitrateWhenNoFilesize covers sources that only report
// TBR, which is the common case for streaming formats.
func TestSelectFormatSpecUsesBitrateWhenNoFilesize(t *testing.T) {
	dur := 100.0
	info := &ytdlp.ExtractedInfo{
		Duration: &dur,
		Formats: []*ytdlp.ExtractedFormat{
			// 2000 Kbit/s * 100 s = 25 MB, under the budget.
			{FormatID: sPtr("v"), VCodec: sPtr("avc1"), Height: f64(720), TBR: f64(2000)},
			{FormatID: sPtr("a"), ACodec: sPtr("mp4a"), TBR: f64(128)},
		},
	}
	if got := selectFormatSpec(info); got != "v+a" {
		t.Errorf("selectFormatSpec = %q, want %q", got, "v+a")
	}
}

// TestSelectFormatSpecIgnoresNoneCodec asserts yt-dlp's literal "none" codec
// marker is treated as absent, not as a real stream.
func TestSelectFormatSpecIgnoresNoneCodec(t *testing.T) {
	dur := 60.0
	info := &ytdlp.ExtractedInfo{
		Duration: &dur,
		Formats: []*ytdlp.ExtractedFormat{
			{FormatID: sPtr("140"), VCodec: sPtr("none"), ACodec: sPtr("mp4a.40.2"), FileSize: iPtr(1_000_000)},
			{FormatID: sPtr("137"), VCodec: sPtr("avc1"), ACodec: sPtr("none"), Height: f64(1080), FileSize: iPtr(20_000_000)},
		},
	}
	if got := selectFormatSpec(info); got != "137+140" {
		t.Errorf("selectFormatSpec = %q, want %q", got, "137+140")
	}
}

func TestCodecPredicates(t *testing.T) {
	cases := []struct {
		name       string
		f          *ytdlp.ExtractedFormat
		h264, aac  bool
		hasV, hasA bool
	}{
		{"nil", nil, false, false, false, false},
		{"avc+mp4a", &ytdlp.ExtractedFormat{VCodec: sPtr("avc1.42001E"), ACodec: sPtr("mp4a.40.2")}, true, true, true, true},
		{"vp9", &ytdlp.ExtractedFormat{VCodec: sPtr("vp9"), ACodec: sPtr("none")}, false, false, true, false},
		{"hevc", &ytdlp.ExtractedFormat{VCodec: sPtr("hev1.1.6")}, false, false, true, false},
		{"opus-only", &ytdlp.ExtractedFormat{VCodec: sPtr("none"), ACodec: sPtr("opus")}, false, false, false, true},
		{"h264-literal", &ytdlp.ExtractedFormat{VCodec: sPtr("h264"), ACodec: sPtr("aac")}, true, true, true, true},
	}
	for _, c := range cases {
		if got := isH264(c.f); got != c.h264 {
			t.Errorf("%s: isH264 = %v, want %v", c.name, got, c.h264)
		}
		if got := isAAC(c.f); got != c.aac {
			t.Errorf("%s: isAAC = %v, want %v", c.name, got, c.aac)
		}
		if got := hasVideoCodec(c.f); got != c.hasV {
			t.Errorf("%s: hasVideoCodec = %v, want %v", c.name, got, c.hasV)
		}
		if got := hasAudioCodec(c.f); got != c.hasA {
			t.Errorf("%s: hasAudioCodec = %v, want %v", c.name, got, c.hasA)
		}
	}
}

func TestFormatBytesAndHeight(t *testing.T) {
	dur := 100.0

	if got := formatBytes(nil, dur); got != 0 {
		t.Errorf("formatBytes(nil) = %d, want 0", got)
	}
	// Exact size wins over the estimate.
	f := &ytdlp.ExtractedFormat{FileSize: iPtr(1234), TBR: f64(2000)}
	if got := formatBytes(f, dur); got != 1234 {
		t.Errorf("formatBytes = %d, want the exact 1234", got)
	}
	// Approximate size is next.
	f = &ytdlp.ExtractedFormat{FileSizeApprox: iPtr(4321)}
	if got := formatBytes(f, dur); got != 4321 {
		t.Errorf("formatBytes = %d, want the approx 4321", got)
	}
	// Bitrate: 2000 Kbit/s * 100 s / 8 = 25_000_000 bytes.
	f = &ytdlp.ExtractedFormat{TBR: f64(2000)}
	if got := formatBytes(f, dur); got != 25_000_000 {
		t.Errorf("formatBytes from TBR = %d, want 25000000", got)
	}
	// No duration means the bitrate cannot be turned into bytes.
	if got := formatBytes(f, 0); got != 0 {
		t.Errorf("formatBytes with no duration = %d, want 0", got)
	}

	if got := formatHeight(&ytdlp.ExtractedFormat{Height: f64(720)}); got != 720 {
		t.Errorf("formatHeight = %d, want 720", got)
	}
	if got := formatHeight(&ytdlp.ExtractedFormat{}); got != 0 {
		t.Errorf("formatHeight unknown = %d, want 0", got)
	}
}

// TestLiveSelectFormatSpecMatchesDownload confirms the pre-selected format is
// one yt-dlp will actually download, which is the assumption the fast path runs
// on. A spec rejected here would silently fall through to the slow cascade.
func TestLiveSelectFormatSpecMatchesDownload(t *testing.T) {
	liveEnabled(t)
	cfg := testConfig()

	for _, u := range []string{
		"https://www.youtube.com/shorts/DXGyKxzUZBU",
		"https://vt.tiktok.com/ZS4sgaFqm",
	} {
		info, err := extractInfo(context.Background(), u, cfg)
		if err != nil {
			t.Errorf("%s extract: %v", u, err)
			continue
		}
		spec := selectFormatSpec(info)
		if spec == "" {
			t.Logf("%s: no pre-selectable format (sizes unknown), cascade will handle it", u)
			continue
		}

		dir := t.TempDir()
		path, err := downloadByFormat(context.Background(), u, dir, spec, cfg)
		if err != nil {
			t.Errorf("%s: pre-selected spec %q failed to download: %v", u, spec, err)
			continue
		}
		fi, _ := os.Stat(path)
		data, _ := os.ReadFile(path)
		hasAudio := strings.Contains(string(data), "soun")

		t.Logf("%s spec=%s size=%.1fMB audio=%v", shortName(u), spec,
			float64(fi.Size())/1048576, hasAudio)

		if fi.Size() >= maxFileBytes {
			t.Errorf("%s: pre-selected file is %d bytes, over the upload limit", u, fi.Size())
		}
		if !hasAudio {
			t.Errorf("%s: pre-selected file has no audio track", u)
		}
	}
}

func shortName(u string) string {
	if i := strings.Index(u, "//"); i >= 0 {
		u = u[i+2:]
	}
	if i := strings.Index(u, "/"); i > 0 {
		u = u[:i]
	}
	return u
}

// TestLiveYouTubeResolution is the regression guard for the quality bug: the
// pinned player_client used to collapse YouTube to 360p. It asserts that
// metadata extraction now sees high-resolution separate streams.
//
// Requires network plus a valid cookies file, hence the opt-in env var.
func TestLiveYouTubeResolution(t *testing.T) {
	liveEnabled(t)

	info, err := extractInfo(context.Background(), "https://www.youtube.com/shorts/DXGyKxzUZBU", testConfig())
	if err != nil {
		t.Skipf("extract failed (cookies or network): %v", err)
	}

	// The default client exposes dozens of formats; the broken pin exposed ~5.
	if len(info.Formats) < 10 {
		t.Errorf("only %d formats exposed — a player_client pin is likely clamping quality", len(info.Formats))
	}

	var maxH float64
	for _, f := range info.Formats {
		if f.Height != nil && *f.Height > maxH {
			maxH = *f.Height
		}
	}
	if maxH < 480 {
		t.Errorf("max available height = %.0fp, expected >= 480p for this short", maxH)
	}
	t.Logf("formats=%d maxHeight=%.0fp", len(info.Formats), maxH)
}

// TestLiveTikTokPhotoPost verifies the scraper against the real post, since
// yt-dlp has no extractor for these and the page structure can change.
func TestLiveTikTokPhotoPost(t *testing.T) {
	liveEnabled(t)

	dir := t.TempDir()
	post, err := scrapeTikTokPhotoPost(context.Background(),
		"https://www.tiktok.com/@lil.mill_000/photo/7633745141773257991", dir, "")
	if err != nil {
		t.Skipf("scrape failed (page structure may have changed): %v", err)
	}

	if len(post.Images) == 0 {
		t.Fatal("no images scraped")
	}
	if len(post.Images) > maxPhotoPostImages {
		t.Errorf("scraped %d images, want at most %d", len(post.Images), maxPhotoPostImages)
	}

	for _, p := range post.Images {
		fi, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if fi.Size() == 0 {
			t.Errorf("empty image at %s", p)
		}
		// A JPEG starts with FF D8; catching an HTML error page saved as .jpg.
		head := make([]byte, 2)
		f, err := os.Open(p)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = f.Read(head)
		_ = f.Close()
		if head[0] != 0xFF || head[1] != 0xD8 {
			t.Errorf("%s is not a JPEG (magic % x)", filepath.Base(p), head)
		}
	}
	t.Logf("scraped %d images, title=%.60q", len(post.Images), post.Title)
}

// TestHasVideoStream covers the audio-only detection that routes a TikTok photo
// post shared as a /video/ link to the scraper instead of uploading an mp3 as a
// "video" with a black screen.
func TestHasVideoStream(t *testing.T) {
	if hasVideoStream(nil) {
		t.Error("nil info should report no video stream")
	}

	// No format list means unknown, which must not be refused: simulate mode does
	// not always enumerate formats.
	if !hasVideoStream(&ytdlp.ExtractedInfo{}) {
		t.Error("empty format list should be treated as unknown and allowed")
	}

	// Audio-only catalog — the photo-post-as-video case.
	audioOnly := &ytdlp.ExtractedInfo{Formats: []*ytdlp.ExtractedFormat{
		{FormatID: sPtr("0"), VCodec: sPtr("none"), ACodec: sPtr("mp4a.40.2")},
		{FormatID: sPtr("1"), VCodec: sPtr("none"), ACodec: sPtr("opus")},
	}}
	if hasVideoStream(audioOnly) {
		t.Error("audio-only media should report no video stream")
	}

	// A real video catalog.
	withVideo := &ytdlp.ExtractedInfo{Formats: []*ytdlp.ExtractedFormat{
		{FormatID: sPtr("0"), VCodec: sPtr("none"), ACodec: sPtr("mp4a.40.2")},
		{FormatID: sPtr("1"), VCodec: sPtr("avc1.42001E"), ACodec: sPtr("mp4a.40.2")},
	}}
	if !hasVideoStream(withVideo) {
		t.Error("media with an avc1 stream should report a video stream")
	}
}

// TestLivePhotoPostAsVideoURL verifies the routing fix on a real URL: a TikTok
// photo post addressed as /video/ extracts audio only, so it must be detected
// and handed to the scraper.
func TestLivePhotoPostAsVideoURL(t *testing.T) {
	liveEnabled(t)

	info, err := extractInfo(context.Background(),
		"https://www.tiktok.com/@lil.mill_000/video/7633745141773257991", testConfig())
	if err != nil {
		t.Skipf("extract failed: %v", err)
	}
	if hasVideoStream(info) {
		t.Log("tiktok now exposes a video stream for this photo post; routing fix is a no-op")
		return
	}
	t.Logf("confirmed audio-only (%d formats), scraper fallback will handle it", len(info.Formats))
}

// TestProxyClientHandlesBareCredentialsForm is the regression guard for a bug
// where url.Parse was called on the proxy string directly. The documented
// configuration form "login:pass@host:port" has no scheme, so url.Parse read
// "login" as the scheme and left Host empty — producing a client that silently
// never used the proxy.
func TestProxyClientHandlesBareCredentialsForm(t *testing.T) {
	client, err := proxyClient("login:pass@1.2.3.4:8080")
	if err != nil {
		t.Fatalf("bare credentials form should parse: %v", err)
	}
	tr, ok := client.Transport.(*http.Transport)
	if !ok || tr == nil {
		t.Fatal("client has no *http.Transport")
	}
	if tr.Proxy == nil {
		t.Fatal("transport has no Proxy func; the proxy would be silently ignored")
	}

	u, err := tr.Proxy(&http.Request{URL: &url.URL{Scheme: "https", Host: "www.tiktok.com"}})
	if err != nil {
		t.Fatalf("Proxy(): %v", err)
	}
	if u == nil {
		t.Fatal("Proxy() returned nil, so requests would go direct")
	}
	if u.Host != "1.2.3.4:8080" {
		t.Errorf("proxy target = %q, want %q", u.Host, "1.2.3.4:8080")
	}
	// Credentials must survive parsing, or the proxy answers 407.
	if u.User == nil || u.User.Username() != "login" {
		t.Errorf("proxy credentials lost during parse: user=%v", u.User)
	}
	if pass, ok := u.User.Password(); !ok || pass != "pass" {
		t.Error("proxy password lost during parse")
	}
}

// TestProxyClientRejectsInvalidInsteadOfGoingDirect guards the second half of
// the fix: a malformed proxy must be an error, not a silent fallback to a direct
// connection. Going direct on a network where the destination is blocked looks
// exactly like a TikTok outage.
func TestProxyClientRejectsInvalidInsteadOfGoingDirect(t *testing.T) {
	if _, err := proxyClient("not a proxy url at all"); err == nil {
		t.Error("a malformed proxy specification must return an error")
	}
}

func TestProxyClientEmptyMeansDirect(t *testing.T) {
	client, err := proxyClient("")
	if err != nil {
		t.Fatalf("empty proxy should be valid (direct): %v", err)
	}
	if client != http.DefaultClient {
		t.Error("empty proxy should yield http.DefaultClient")
	}
}

// TestIsFatalFormatError covers the classification that lets a refused format
// short-circuit the resolution ladder. A 403 does not change when the resolution
// changes, so retrying it at 1080/720/480 only burns the per-video deadline —
// measured at roughly a minute per rung, which is how a working progressive
// fallback ended up never being reached.
func TestIsFatalFormatError(t *testing.T) {
	fatal := []string{
		"exit code 1: exit status 1\n\nERROR: unable to download video data: HTTP Error 403: Forbidden",
		"ERROR: [youtube] PAniXPAaAFg: Requested format is not available. Use --list-formats for a list of available formats",
		"ERROR: [youtube] x: Sign in to confirm you're not a bot.",
		"ERROR: No video formats found",
		"http error 403",
		"FORBIDDEN",
	}
	for _, msg := range fatal {
		if !isFatalFormatError(errors.New(msg)) {
			t.Errorf("expected fatal (no retry): %q", truncateMsg(msg, 70))
		}
	}

	// Transient failures must stay retryable, or a flaky network would make the
	// bot give up on videos it could have fetched.
	retryable := []string{
		"read timeout",
		"connection reset by peer",
		"TLS handshake timeout",
		"temporary failure in name resolution",
		"HTTP Error 500: Internal Server Error",
		"HTTP Error 503: Service Unavailable",
		"HTTP Error 429: Too Many Requests",
		"unable to download webpage",
	}
	for _, msg := range retryable {
		if isFatalFormatError(errors.New(msg)) {
			t.Errorf("expected retryable, got fatal: %q", truncateMsg(msg, 70))
		}
	}

	if isFatalFormatError(nil) {
		t.Error("nil error must not be fatal")
	}
}

func truncateMsg(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
