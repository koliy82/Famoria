package link

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"famoria/internal/pkg/html"
	"famoria/internal/pkg/i18n"

	"github.com/lrstanley/go-ytdlp"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const (
	maxDurationSeconds = 30 * 60      // skip videos longer than 30 minutes
	maxFileBytes       = 50 * 1 << 20 // Telegram bot upload limit: 50 MB
	perVideoTimeout    = 10 * 60      // 10 minutes per video processing

	// captionLimit is Telegram's hard limit for a video caption. Titles longer
	// than this are truncated, or the whole send fails with 400.
	captionLimit = 1024
)

// resolutionLadder is walked only when a file exceeds the 50 MB upload limit.
// The first attempt is always uncapped, so quality is never reduced unless the
// upload limit forces it.
var resolutionLadder = []int{0, 1080, 720, 480}

// videoStrategy is one way of asking yt-dlp for the media, in preference order.
type videoStrategy struct {
	name   string
	format string
}

// telegramFormatSort prefers H.264 + AAC, which is what Telegram plays with
// sound and without re-encoding. Other codecs (HEVC, VP9, Opus) arrive muted.
//
// Both codec namings are listed because sources disagree: TikTok reports
// "h264"/"aac" while YouTube reports "avc1.42001E"/"mp4a.40.2". Naming only one
// family silently loses the other.
//
// When height > 0 the "res:height" prefix makes yt-dlp prefer the closest
// resolution to that height; when height == 0 "res" alone prefers the highest,
// which is what keeps quality at its ceiling instead of clamping to 720p.
func telegramFormatSort(height int) string {
	res := "res"
	if height > 0 {
		res = fmt.Sprintf("res:%d", height)
	}
	return res + ",vcodec:avc1:h264,acodec:mp4a:aac,ext:mp4:m4a"
}

// combinedSelector matches a progressive (already video+audio) stream in the
// codec naming each platform actually uses.
//
// This must list avc1/mp4a before h264/aac: YouTube publishes its codecs as
// "avc1.42001E" and "mp4a.40.2", so a selector written only as [vcodec^=h264]
// matches nothing on YouTube and yt-dlp answers "Requested format is not
// available". TikTok, conversely, reports literally "h264" and "aac". Both
// spellings are therefore required.
const combinedSelector = "best[vcodec^=avc1][acodec^=mp4a]/" +
	"best[vcodec^=h264][acodec^=aac]/" +
	"best[acodec^=aac]/" +
	"best"

// Strategies differ per platform because their catalogs differ:
//
//   - YouTube serves high resolutions only as separate DASH video+audio
//     streams. Its progressive (combined) H.264+AAC stream stops at 360-480p, so
//     asking for the combined stream first would clamp quality. Merging separate
//     streams is tried first, with the combined stream kept as a fallback because
//     DASH downloads can fail with HTTP 403 when YouTube demands a PO token.
//   - TikTok and Instagram serve only combined streams. Asking yt-dlp to merge
//     separate video+audio there produces a file with no audio track at all, so
//     the combined stream must come first. Their H.264 variant may also exist
//     only below the resolution of their HEVC one, so no hard height filter is
//     ever applied — resolution is steered through FormatSort, which prefers
//     rather than excludes.
//
// When ffmpeg is unavailable the merge strategies are dropped entirely. A "+"
// spec cannot be muxed without it, and yt-dlp does not report that as an error:
// it writes both halves as separate files and exits 0, after which the largest
// .mp4 — the silent video-only half — is what gets picked. Better to serve a
// lower-resolution progressive stream with sound than a 1080p video without it.
func videoStrategies(isYT bool) []videoStrategy {
	combined := videoStrategy{
		name:   "combined-h264",
		format: combinedSelector,
	}

	if !ffmpegAvailable() {
		// Progressive only: these formats already contain both tracks, so no
		// muxing is needed.
		return []videoStrategy{combined}
	}

	dash := videoStrategy{
		name:   "dash-h264",
		format: "bestvideo[vcodec^=avc1]+bestaudio[acodec^=mp4a]/bestvideo+bestaudio",
	}
	fallback := videoStrategy{
		name:   "best",
		format: "bestvideo+bestaudio/best",
	}

	if isYT {
		return []videoStrategy{dash, combined, fallback}
	}
	return []videoStrategy{combined, dash, fallback}
}

// isYouTube reports whether the URL points at a YouTube host. YouTube needs
// cookies because it blocks unauthenticated (bot) access from datacenter IPs.
func isYouTube(rawURL string) bool {
	host, ok := urlHost(rawURL)
	if !ok {
		return false
	}
	switch host {
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com",
		"youtu.be", "www.youtu.be":
		return true
	}
	return strings.HasSuffix(host, ".youtube.com")
}

// isTikTok reports whether the URL points at a TikTok host, including the vt.
// and vm. short-link forms.
func isTikTok(rawURL string) bool {
	host, ok := urlHost(rawURL)
	return ok && strings.HasSuffix(host, "tiktok.com")
}

// urlHost returns the lowercased hostname (without port) of a URL.
func urlHost(rawURL string) (string, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", false
	}
	return host, true
}

// Proxy mode constants, matching the PROXY_ENABLE config values.
const (
	proxyModeFalse   = "false"   // proxy disabled (default)
	proxyModeYouTube = "youtube" // proxy only for YouTube URLs
	proxyModeAll     = "true"    // proxy for all URLs
)

// ytConfig bundles the optional yt-dlp authentication/network parameters.
type ytConfig struct {
	cookiesFile string // Netscape cookies file, applied to YouTube only.
	proxy       string // HTTP/SOCKS proxy URL.
	proxyMode   string // when to apply proxy: false/youtube/true.
}

// useProxy reports whether the proxy applies to this URL under the configured
// mode.
func (c ytConfig) useProxy(rawURL string) bool {
	if c.proxy == "" {
		return false
	}
	switch c.proxyMode {
	case proxyModeAll:
		return true
	case proxyModeYouTube:
		return isYouTube(rawURL)
	default: // proxyModeFalse or anything unrecognized
		return false
	}
}

// applyParams attaches cookies and the proxy where they apply.
//
// Note there is deliberately no player_client override here. Pinning
// youtube:player_client to web,mweb,android cuts the format list from ~35 down
// to 5 and caps every video at 360p, because only the low-resolution
// progressive stream is exposed through those clients. The default client set
// is both unblocked (cookies handle the bot check) and far higher quality.
func applyParams(cmd *ytdlp.Command, rawURL string, cfg ytConfig) *ytdlp.Command {
	if isYouTube(rawURL) && cfg.cookiesFile != "" {
		cmd = cmd.Cookies(cfg.cookiesFile)
	}
	if cfg.useProxy(rawURL) {
		cmd = cmd.Proxy(cfg.proxy)
	}
	return cmd
}

// extractInfo queries yt-dlp for metadata without downloading the media.
// Returns the first entry, or an error when the URL is not supported media.
func extractInfo(ctx context.Context, rawURL string, cfg ytConfig) (*ytdlp.ExtractedInfo, error) {
	cmd := ytdlp.New().
		DumpJSON().
		Simulate().
		NoPlaylist().
		NoWarnings().
		Quiet().
		NoColors()
	cmd = applyParams(cmd, rawURL, cfg)

	r, err := cmd.Run(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	info, err := r.GetExtractedInfo()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, fmt.Errorf("no extractable media info for %s", rawURL)
	}
	return info[0], nil
}

// downloadVideo downloads the media with one strategy at one resolution target
// into dir and returns the path to the produced mp4.
//
// MergeOutputFormat (a remux) is used rather than RecodeVideo (a re-encode):
// re-encoding reliably stripped the audio track from already-combined streams
// such as TikTok's, and it is also what degraded quality unnecessarily.
func downloadVideo(ctx context.Context, rawURL, dir string, s videoStrategy, height int, cfg ytConfig) (string, error) {
	cmd := ytdlp.New().
		Format(s.format).
		FormatSort(telegramFormatSort(height)).
		MergeOutputFormat("mp4").
		NoPlaylist().
		NoContinue().
		NoPart().
		ForceOverwrites().
		NoProgress().
		Quiet().
		NoWarnings().
		NoColors().
		Output(filepath.Join(dir, "video.%(ext)s"))
	cmd = applyParams(cmd, rawURL, cfg)

	if _, err := cmd.Run(ctx, rawURL); err != nil {
		return "", err
	}
	return findMP4(dir)
}

// findMP4 returns the largest .mp4 in dir. yt-dlp can leave intermediate
// artifacts behind, and the merged result is always the largest file.
func findMP4(dir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.mp4"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no mp4 produced in %s", dir)
	}
	var best string
	var bestSize int64 = -1
	for _, m := range matches {
		fi, err := os.Stat(m)
		if err != nil {
			continue
		}
		if fi.Size() > bestSize {
			bestSize = fi.Size()
			best = m
		}
	}
	if best == "" {
		return "", fmt.Errorf("no readable mp4 in %s", dir)
	}
	return best, nil
}

// clearDir removes every entry inside dir while keeping the directory itself, so
// it stays usable for the next download attempt. Errors are ignored: the worst
// case is a leftover file, which findMP4 then resolves by size.
func clearDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		_ = os.RemoveAll(filepath.Join(dir, e.Name()))
	}
}

// buildCaption renders the text placed under the media: the source title when
// available, the link back to where it was posted, and the bot attribution.
//
// The title is HTML-escaped and the whole caption is clamped to Telegram's 1024
// character limit. Over-long titles are shortened rather than dropped, because
// a caption over the limit makes Telegram reject the entire send.
func buildCaption(lang i18n.Lang, title, sourceURL string) string {
	link := html.Link(sourceURL, i18n.T(lang, i18n.KeyVideoOriginal))
	title = strings.TrimSpace(title)

	if title == "" {
		return i18n.T(lang, i18n.KeyVideoCaptionNoTitle, link)
	}

	caption := i18n.T(lang, i18n.KeyVideoCaption, html.Escape(title), link)
	if len(caption) <= captionLimit {
		return caption
	}

	// Shorten the title until the rendered caption fits. Only the title is cut:
	// the source link and the attribution must survive, since losing the link
	// would leave the repost unattributed.
	for runes := []rune(title); len(runes) > 1; {
		// Drop a chunk at a time; binary search would be overkill for a caption.
		step := len(runes) / 8
		if step < 1 {
			step = 1
		}
		runes = runes[:len(runes)-step]

		candidate := i18n.T(lang, i18n.KeyVideoCaption, html.Escape(string(runes))+"…", link)
		if len(candidate) <= captionLimit {
			return candidate
		}
	}

	return i18n.T(lang, i18n.KeyVideoCaptionNoTitle, link)
}

// sendVideo uploads a local video file and returns the sent message.
//
// The file is attached under a clean "video.mp4" name: telego passes
// NamedReader.Name() straight into the multipart filename, and *os.File
// reports a full Windows/temp path there, which makes Telegram misclassify the
// upload and drop the audio. Geometry from yt-dlp is passed along so Telegram
// does not have to guess it.
func sendVideo(ctx context.Context, bot *telego.Bot, chatID int64, replyTo int, path string, params videoMeta) (*telego.Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &telego.SendVideoParams{
		ChatID:            tu.ID(chatID),
		Video:             tu.FileFromReader(f, "video.mp4"),
		Caption:           params.Caption,
		ParseMode:         telego.ModeHTML,
		SupportsStreaming: true,
		Duration:          params.Duration,
		Width:             params.Width,
		Height:            params.Height,
	}
	if replyTo != 0 {
		p.ReplyParameters = &telego.ReplyParameters{
			MessageID:                replyTo,
			AllowSendingWithoutReply: true,
		}
	}
	return bot.SendVideo(ctx, p)
}

// videoMeta carries the caption and geometry for one send.
type videoMeta struct {
	Caption  string
	Duration int
	Width    int
	Height   int
}

// metaFromInfo derives send parameters from yt-dlp metadata.
func metaFromInfo(info *ytdlp.ExtractedInfo, caption string) videoMeta {
	m := videoMeta{Caption: caption}
	if info == nil {
		return m
	}
	if info.Duration != nil {
		m.Duration = int(*info.Duration)
	}
	if info.Width != nil {
		m.Width = int(*info.Width)
	}
	if info.Height != nil {
		m.Height = int(*info.Height)
	}
	return m
}

// estimatedSizeBudget is the upload limit minus headroom. Sizes derived from
// bitrate are estimates, so leaving room avoids picking a combination that then
// fails the exact post-download check.
const estimatedSizeBudget = int64(46 * 1 << 20)

// ffmpegAvailable reports whether ffmpeg can be executed.
//
// This matters more than it looks: yt-dlp does NOT fail when ffmpeg is missing.
// It downloads the video and audio as separate files and reports success, and a
// merge format such as "137+140-0" silently yields the video-only half. That
// produces a "successful" send with no sound — the exact failure this guards.
func ffmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// selectFormatSpec picks a Telegram-compatible format combination whose size
// fits the upload limit, using metadata that extractInfo already fetched.
//
// Without this the strategy cascade would download a whole file only to discover
// it exceeds 50 MB, then download another — up to a dozen attempts per link,
// which cannot finish inside the processing deadline for longer videos.
// Estimating from filesize/bitrate up front makes the common case one download.
//
// It returns an empty string when no format reports a usable size, in which case
// the caller falls back to the strategy cascade.
func selectFormatSpec(info *ytdlp.ExtractedInfo) string {
	if info == nil {
		return ""
	}

	var duration float64
	if info.Duration != nil {
		duration = *info.Duration
	}

	// Split the catalog the way yt-dlp would: audio-only, video-only and
	// progressive (already-muxed) formats.
	var audio, video, progressive []*ytdlp.ExtractedFormat
	for _, f := range info.Formats {
		switch hasV, hasA := hasVideoCodec(f), hasAudioCodec(f); {
		case hasV && hasA:
			progressive = append(progressive, f)
		case hasV:
			video = append(video, f)
		case hasA:
			audio = append(audio, f)
		}
	}

	type candidate struct {
		spec   string
		height int
		bytes  int64
	}
	var best *candidate

	consider := func(c candidate) {
		// Only accept combinations with a known size: without one the upload
		// limit cannot be checked before downloading, so leave those to the
		// cascade.
		if c.spec == "" || c.bytes <= 0 || c.bytes > estimatedSizeBudget {
			return
		}
		if best == nil {
			cp := c
			best = &cp
			return
		}
		// Prefer the tallest picture; within one height a larger file means a
		// higher bitrate.
		if c.height > best.height || (c.height == best.height && c.bytes > best.bytes) {
			cp := c
			best = &cp
		}
	}

	for _, p := range progressive {
		if !isH264(p) || !isAAC(p) {
			continue
		}
		consider(candidate{
			spec:   infoString(p.FormatID),
			height: formatHeight(p),
			bytes:  formatBytes(p, duration),
		})
	}

	// Pair each H.264 video format with the best AAC audio track, which is what
	// a merge actually produces. This requires ffmpeg; without it the pair is
	// never muxed and the result is a silent video, so skip it entirely.
	if ffmpegAvailable() {
		var bestAudio *ytdlp.ExtractedFormat
		for _, a := range audio {
			if !isAAC(a) {
				continue
			}
			if bestAudio == nil || formatBytes(a, duration) > formatBytes(bestAudio, duration) {
				bestAudio = a
			}
		}
		if bestAudio != nil {
			audioID := infoString(bestAudio.FormatID)
			audioBytes := formatBytes(bestAudio, duration)
			if audioID != "" && audioBytes > 0 {
				for _, v := range video {
					if !isH264(v) {
						continue
					}
					vb := formatBytes(v, duration)
					if vb <= 0 {
						continue
					}
					consider(candidate{
						spec:   infoString(v.FormatID) + "+" + audioID,
						height: formatHeight(v),
						bytes:  vb + audioBytes,
					})
				}
			}
		}
	}

	if best == nil {
		return ""
	}
	return best.spec
}

// formatBytes estimates a format's size, preferring an exact figure over one
// derived from bitrate. Returns 0 when neither is available.
func formatBytes(f *ytdlp.ExtractedFormat, durationSec float64) int64 {
	if f == nil {
		return 0
	}
	if f.FileSize != nil && *f.FileSize > 0 {
		return int64(*f.FileSize)
	}
	if f.FileSizeApprox != nil && *f.FileSizeApprox > 0 {
		return int64(*f.FileSizeApprox)
	}
	// TBR is the average audio+video bitrate in Kbit/s.
	if f.TBR != nil && *f.TBR > 0 && durationSec > 0 {
		return int64(*f.TBR * durationSec * 1000 / 8)
	}
	return 0
}

// formatHeight returns a format's pixel height, or 0 when unknown.
func formatHeight(f *ytdlp.ExtractedFormat) int {
	if f == nil || f.Height == nil {
		return 0
	}
	return int(*f.Height)
}

// hasVideoCodec reports whether the format carries a video stream. yt-dlp marks
// audio-only formats with the literal codec "none".
func hasVideoCodec(f *ytdlp.ExtractedFormat) bool {
	if f == nil || f.VCodec == nil {
		return false
	}
	c := strings.ToLower(*f.VCodec)
	return c != "" && c != "none"
}

// hasAudioCodec reports whether the format carries an audio stream.
func hasAudioCodec(f *ytdlp.ExtractedFormat) bool {
	if f == nil || f.ACodec == nil {
		return false
	}
	c := strings.ToLower(*f.ACodec)
	return c != "" && c != "none"
}

// isH264 reports whether the video codec is H.264/AVC, which Telegram plays
// natively. HEVC and VP9 arrive as video many clients cannot play.
func isH264(f *ytdlp.ExtractedFormat) bool {
	if !hasVideoCodec(f) {
		return false
	}
	c := strings.ToLower(*f.VCodec)
	return strings.HasPrefix(c, "avc1") || strings.HasPrefix(c, "avc3") || c == "h264"
}

// isAAC reports whether the audio codec is AAC, the other half of what Telegram
// plays with sound. Opus inside an mp4 container plays silent.
func isAAC(f *ytdlp.ExtractedFormat) bool {
	if !hasAudioCodec(f) {
		return false
	}
	c := strings.ToLower(*f.ACodec)
	return strings.HasPrefix(c, "mp4a") || c == "aac"
}

// downloadByFormat downloads one exact format specification into dir. Used when
// selectFormatSpec pre-picked a combination that fits the size budget.
func downloadByFormat(ctx context.Context, rawURL, dir, spec string, cfg ytConfig) (string, error) {
	cmd := ytdlp.New().
		Format(spec).
		MergeOutputFormat("mp4").
		NoPlaylist().
		NoContinue().
		NoPart().
		ForceOverwrites().
		NoProgress().
		Quiet().
		NoWarnings().
		NoColors().
		Output(filepath.Join(dir, "video.%(ext)s"))
	cmd = applyParams(cmd, rawURL, cfg)

	if _, err := cmd.Run(ctx, rawURL); err != nil {
		return "", err
	}
	return findMP4(dir)
}

// infoString reads a possibly-nil yt-dlp string field.
func infoString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// hasVideoStream reports whether the extracted media contains any video track.
//
// Some URLs yield audio only — a TikTok photo post shared as a /video/ link, for
// instance, exposes just its soundtrack. Uploading that as a video would send a
// black screen with sound, so the caller needs to know before committing to the
// video path.
//
// An empty format list is treated as "unknown" and reported true: yt-dlp does not
// always enumerate formats in simulate mode, and refusing those would skip videos
// that download fine.
func hasVideoStream(info *ytdlp.ExtractedInfo) bool {
	if info == nil {
		return false
	}
	if len(info.Formats) == 0 {
		return true
	}
	for _, f := range info.Formats {
		if hasVideoCodec(f) {
			return true
		}
	}
	return false
}

// hasAudioStream reports whether the source offers any audio track.
//
// Note the "unknown" case is deliberately the opposite of hasVideoStream. This
// predicate only gates a refusal to send: an absent audio track is treated as a
// failure solely when the source provably had one. So an empty format list —
// which simulate mode sometimes produces — must return false, otherwise a video
// whose formats were never enumerated would be rejected for lacking audio it may
// well have.
func hasAudioStream(info *ytdlp.ExtractedInfo) bool {
	if info == nil || len(info.Formats) == 0 {
		return false
	}
	for _, f := range info.Formats {
		if hasAudioCodec(f) {
			return true
		}
	}
	return false
}
