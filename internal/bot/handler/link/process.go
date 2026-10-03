package link

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"famoria/internal/database/mongo/repositories/media_cache"
	"famoria/internal/pkg/common/normalize"
	"famoria/internal/pkg/i18n"

	"github.com/lrstanley/go-ytdlp"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

// sendRequest describes one link that should end up as media in a chat.
type sendRequest struct {
	bot   *telego.Bot
	cache media_cache.Repository
	log   *zap.Logger
	cfg   ytConfig

	chatID      int64
	replyTo     int
	originalURL string
	// mediaURL is the address to fetch media from, set when a share link had to be
	// resolved to the canonical post. Empty means "same as originalURL".
	mediaURL string
	lang     i18n.Lang
}

// fetchURL returns the URL to hand to yt-dlp or the scraper.
func (r sendRequest) fetchURL() string {
	if r.mediaURL != "" {
		return r.mediaURL
	}
	return r.originalURL
}

// process resolves a link into media sent to the chat.
//
// The order matters: a cached entry avoids any network download at all, a
// TikTok photo post needs scraping rather than yt-dlp, and only then does the
// video path run. It returns true when media reached the chat, which is what
// lets the caller delete the original link message.
func process(ctx context.Context, r sendRequest) bool {
	key := normalize.URL(r.originalURL)

	if key != "" && r.cache != nil {
		if entry, resolved := r.resolveCache(key); entry != nil {
			if r.replay(ctx, entry, key) {
				return true
			}
			// A stale file_id: drop the entry that pointed at it so the next
			// share re-downloads instead of failing again.
			r.log.Info("media: cached entry rejected, dropping",
				zap.String("key", resolved), zap.String("alias", key))
			_ = r.cache.Delete(resolved)
			if resolved != key {
				_ = r.cache.Delete(key)
			}
		}
	}

	// A TikTok share link (vt./vm.) carries no /photo/ or /video/ segment, so the
	// post type is unknown until the redirect is followed. Resolving here is what
	// lets a shared photo post reach the scraper instead of yt-dlp, which rejects
	// photo URLs as "Unsupported URL".
	target := r.originalURL
	if isTikTokShortLink(target) {
		proxySpec := ""
		if r.cfg.useProxy(target) {
			proxySpec = r.cfg.proxy
		}
		resolved, ok := resolveTikTokShortLink(ctx, target, proxySpec)
		if ok && resolved != target {
			r.log.Info("link: resolved TikTok share link",
				zap.String("from", target), zap.String("to", resolved))
			target = resolved
			// Cache under the canonical post, so every short link that resolves here
			// hits the same entry instead of downloading again per share.
			if canonical := normalize.URL(resolved); canonical != "" {
				key = canonical
			}
		}
	}

	if isTikTokPhotoPost(target) {
		return processPhotoPost(ctx, r.withMediaURL(target), key)
	}
	return processVideo(ctx, r.withMediaURL(target), key)
}

// withMediaURL returns a copy whose mediaURL points at the resolved post URL.
//
// mediaURL is what yt-dlp and the scraper fetch; originalURL stays as the link
// the user pasted and is what the caption links back to. They differ for share
// links, where the short form is unusable for photo posts but is the address the
// user should see in the chat.
func (r sendRequest) withMediaURL(mediaURL string) sendRequest {
	r.mediaURL = mediaURL
	return r
}

// resolveCache looks a key up, following one alias hop. It returns the entry and
// the key the entry actually lives under, so a stale entry can be deleted at its
// real address rather than at the alias.
func (r sendRequest) resolveCache(key string) (*media_cache.MediaCache, string) {
	entry, resolved, err := r.cache.Resolve(key)
	if err != nil {
		return nil, key
	}
	if entry == nil || !entry.Replayable() {
		return nil, resolved
	}
	return entry, resolved
}

// replay resends cached media by its stored Telegram file_id, with no download.
// It reports false when Telegram refuses the identifier.
func (r sendRequest) replay(ctx context.Context, entry *media_cache.MediaCache, key string) bool {
	caption := buildCaption(r.lang, entry.Title, r.originalURL)

	switch {
	case entry.IsVideo():
		msg, err := r.bot.SendVideo(ctx, cachedVideoParams(r, entry, caption))
		if err != nil {
			r.log.Info("video: cached file_id rejected",
				zap.String("key", key), zap.Error(err))
			return false
		}
		r.log.Info("video: sent from cache",
			zap.Int64("chat_id", r.chatID), zap.Int("message_id", msg.MessageID),
			zap.String("key", key))
		return true

	case entry.IsPhoto():
		msgs, err := sendPhotos(ctx, r, entry.FileIDs, caption)
		if err != nil {
			r.log.Info("photo: cached file_ids rejected",
				zap.String("key", key), zap.Error(err))
			return false
		}
		r.log.Info("photo: sent from cache",
			zap.Int64("chat_id", r.chatID), zap.Int("count", len(msgs)),
			zap.String("key", key))
		return true
	}
	return false
}

// canonicalKey derives the cache key for the resolved content URL that yt-dlp
// reports, which is the same for every short link pointing at one video.
//
// Short links are regenerated on every share, so keying only on the link the user
// pasted would miss almost every time for TikTok. Keying on the canonical URL is
// what makes repeated shares of one video hit the cache.
func canonicalKey(info *ytdlp.ExtractedInfo) string {
	if info == nil {
		return ""
	}
	return normalize.URL(infoString(info.WebpageURL))
}

// rememberAlias points the key the user actually sent at the canonical key, so
// the next identical link is served with no network call at all.
func (r sendRequest) rememberAlias(key, canonical string) {
	if r.cache == nil || key == "" || canonical == "" || key == canonical {
		return
	}
	if err := r.cache.PutAlias(key, canonical); err != nil {
		r.log.Warn("media: failed to store cache alias",
			zap.String("alias", key), zap.String("target", canonical), zap.Error(err))
	}
}

// cachedVideoParams rebuilds the send parameters for a cached video. Geometry is
// replayed from the stored values because Telegram does not recompute it for a
// file_id send.
func cachedVideoParams(r sendRequest, e *media_cache.MediaCache, caption string) *telego.SendVideoParams {
	p := &telego.SendVideoParams{
		ChatID:            tu.ID(r.chatID),
		Video:             tu.FileFromID(e.FileID),
		Caption:           caption,
		ParseMode:         telego.ModeHTML,
		SupportsStreaming: true,
		Duration:          e.Duration,
		Width:             e.Width,
		Height:            e.Height,
	}
	if r.replyTo != 0 {
		p.ReplyParameters = &telego.ReplyParameters{
			MessageID:                r.replyTo,
			AllowSendingWithoutReply: true,
		}
	}
	return p
}

// processVideo downloads and sends a video, then caches its file_id.
func processVideo(ctx context.Context, r sendRequest, key string) bool {
	// fetchURL is the resolved post address; originalURL stays for logs and the
	// caption, so the chat always sees the link the user actually pasted.
	fetchURL := r.fetchURL()

	info, err := extractInfo(ctx, fetchURL, r.cfg)
	if err != nil {
		// via_proxy is logged because "unable to extract webpage video data" from
		// TikTok usually means the page arrived as an anti-bot stub, which is what
		// a datacenter IP gets. With PROXY_ENABLE=youtube TikTok is not proxied at
		// all, so this field is what separates a proxy misconfiguration from a
		// genuinely broken extractor.
		r.log.Info("video: not supported media or extract failed",
			zap.String("url", fetchURL),
			zap.Bool("via_proxy", r.cfg.useProxy(fetchURL)),
			zap.Error(err))
		return false
	}

	if info.Duration != nil && *info.Duration > maxDurationSeconds {
		r.log.Info("video: skipping, longer than the limit",
			zap.String("url", fetchURL), zap.Float64("seconds", *info.Duration))
		return false
	}

	// A TikTok photo post shared as a /video/ link reaches this path, and yt-dlp
	// extracts only its audio track from those — no video stream at all. Retry as
	// a photo post instead of uploading an mp3 as a "video".
	if !hasVideoStream(info) {
		if isTikTok(fetchURL) {
			r.log.Info("video: no video stream, retrying as a TikTok photo post",
				zap.String("url", fetchURL))
			return processPhotoPost(ctx, r, key)
		}
		r.log.Info("video: no video stream in the extracted media",
			zap.String("url", fetchURL))
		return false
	}

	dir, err := os.MkdirTemp("", "famoria-video-*")
	if err != nil {
		r.log.Error("video: failed to create temp dir", zap.Error(err))
		return false
	}
	// The temp dir holds the only on-disk copy; it is dropped right after the
	// upload, so nothing accumulates between requests.
	defer os.RemoveAll(dir)

	title := infoString(info.Title)
	caption := buildCaption(r.lang, title, r.originalURL)

	// Store the media under the canonical content key, and record the link the
	// user pasted as an alias to it. Short links differ on every share, so this
	// is what lets a second share of the same video hit the cache.
	storeKey := key
	if canonical := canonicalKey(info); canonical != "" {
		storeKey = canonical
	}

	// Fast path: pick one format from metadata that is already known to fit the
	// upload limit. This is a single download instead of the cascade walking the
	// same file over and over, which matters because the whole job has a deadline.
	if spec := selectFormatSpec(info); spec != "" {
		outcome := r.attempt(ctx, dir, info, caption, storeKey, key, title,
			attemptLabel{"selected", spec}, func(ctx context.Context) (string, error) {
				return downloadByFormat(ctx, fetchURL, dir, spec, r.cfg)
			})
		switch outcome {
		case outcomeSent:
			return true
		case outcomeStrategyDead:
			// The chosen format was refused outright. If it was a merge spec, every
			// DASH strategy will be refused the same way, so skip them and go
			// straight to progressive formats rather than spending the deadline on
			// repeats of one refusal.
			if strings.Contains(spec, "+") {
				r.log.Info("video: merge format refused, skipping to progressive strategies",
					zap.String("url", r.originalURL), zap.String("spec", spec))
				return r.cascade(ctx, dir, info, caption, storeKey, key, title, true)
			}
		}
		r.log.Info("video: pre-selected format did not produce a sendable file, falling back to cascade",
			zap.String("url", r.originalURL), zap.String("spec", spec))
	}

	return r.cascade(ctx, dir, info, caption, storeKey, key, title, false)
}

// cascade walks the remaining download strategies and returns whether media
// reached the chat.
//
// skipMerge drops strategies that need ffmpeg to mux separate streams. It is used
// after a merge format was refused outright, when retrying merges only repeats
// the refusal.
func (r sendRequest) cascade(ctx context.Context, dir string, info *ytdlp.ExtractedInfo,
	caption, storeKey, aliasKey, title string, skipMerge bool) bool {

	// Strategies outermost, resolution ladder inside. The first attempt is always
	// uncapped, so quality is only reduced when the upload limit forces it. This
	// path is what handles sources whose metadata hides file sizes.
	fetchURL := r.fetchURL()
	for _, s := range videoStrategies(isYouTube(fetchURL)) {
		if skipMerge && strings.Contains(s.format, "+") {
			r.log.Info("video: skipping merge strategy after it was refused",
				zap.String("strategy", s.name))
			continue
		}
		for _, height := range resolutionLadder {
			strategy, h := s, height
			outcome := r.attempt(ctx, dir, info, caption, storeKey, aliasKey, title,
				attemptLabel{strategy.name, strconv.Itoa(h)}, func(ctx context.Context) (string, error) {
					return downloadVideo(ctx, fetchURL, dir, strategy, h, r.cfg)
				})
			if outcome == outcomeSent {
				return true
			}
			if outcome == outcomeStrategyDead {
				// A 403 or a missing format will not change at another resolution,
				// so abandon this strategy instead of burning the deadline on
				// identical failures.
				r.log.Info("video: strategy cannot succeed, skipping remaining resolutions",
					zap.String("strategy", strategy.name), zap.Int("failed_at_height", h))
				break
			}
		}
	}

	r.log.Info("video: no strategy produced a sendable file",
		zap.String("url", r.originalURL))
	return false
}

// attemptLabel names one try in the logs, so a multi-attempt job stays readable.
type attemptLabel struct {
	strategy string
	detail   string
}

// attemptOutcome distinguishes "try something else" from "stop trying this
// strategy", so a failure that cannot be fixed by lowering resolution does not
// consume the whole deadline repeating itself.
type attemptOutcome int

const (
	// outcomeSent means media reached the chat.
	outcomeSent attemptOutcome = iota
	// outcomeRetry means this attempt failed, but another resolution or strategy
	// may succeed.
	outcomeRetry
	// outcomeStrategyDead means the strategy cannot succeed at any resolution, so
	// the rest of the ladder for it is skipped.
	outcomeStrategyDead
)

// isFatalFormatError reports whether an error means the requested formats are
// unavailable or refused outright, rather than a transient network failure.
//
// HTTP 403 is the case that matters: YouTube refuses DASH streams without a valid
// PO token, and it refuses every DASH format equally. Walking the resolution
// ladder after a 403 therefore repeats one identical failure — measured at about
// a minute each, which consumes most of the per-video deadline before the
// progressive fallback that would have worked is ever reached.
func isFatalFormatError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, marker := range []string{
		"403",
		"forbidden",
		"requested format is not available",
		"no video formats",
		"sign in to confirm",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// attempt runs one download+send cycle: it clears the temp dir, downloads,
// enforces the upload limit on the real file size, verifies the audio track
// survived, sends, and caches the resulting file_id.
//
// storeKey is the canonical key the media is cached under, while aliasKey is the
// key of the link the user actually pasted; the two differ for short links.
//
// The post-download size check stays even though selectFormatSpec estimates
// sizes, because bitrate-derived estimates can be wrong; the exact figure is
// only known once the file exists.
func (r sendRequest) attempt(ctx context.Context, dir string, info *ytdlp.ExtractedInfo,
	caption, storeKey, aliasKey, title string, label attemptLabel, download func(context.Context) (string, error)) attemptOutcome {

	if ctx.Err() != nil {
		r.log.Info("video: processing deadline hit",
			zap.String("url", r.originalURL), zap.String("strategy", label.strategy))
		return outcomeStrategyDead
	}

	// Clear leftovers first: findMP4 picks the largest .mp4 in the directory, so
	// a stale file from a rejected attempt would otherwise win over the fresh one.
	clearDir(dir)

	path, err := download(ctx)
	if err != nil {
		fatal := isFatalFormatError(err)
		r.log.Info("video: download attempt failed",
			zap.String("strategy", label.strategy), zap.String("detail", label.detail),
			zap.String("url", r.originalURL), zap.Bool("strategy_dead", fatal), zap.Error(err))
		if fatal {
			// Lowering the resolution cannot change a 403 or a missing format, so
			// abandon this strategy instead of repeating it per ladder rung.
			return outcomeStrategyDead
		}
		return outcomeRetry
	}

	fi, err := os.Stat(path)
	if err != nil {
		r.log.Info("video: could not stat download",
			zap.String("strategy", label.strategy), zap.Error(err))
		return outcomeRetry
	}
	if fi.Size() >= maxFileBytes {
		r.log.Info("video: over the 50MB upload limit, trying another format",
			zap.String("strategy", label.strategy), zap.String("detail", label.detail),
			zap.Int64("bytes", fi.Size()))
		return outcomeRetry
	}

	// Refuse to upload a file that lost its audio track.
	//
	// yt-dlp does not fail when ffmpeg is missing: for a merge spec like
	// "137+140-0" it writes the video and audio halves as separate files and
	// still exits 0. findMP4 then picks the largest .mp4 — the video-only half —
	// and the send is reported as successful while the user gets a silent video.
	// Comparing against the source metadata is what makes this safe: some videos
	// are genuinely silent, so an absent track is only a failure when the source
	// had one.
	if info != nil && hasAudioStream(info) {
		probe, err := probeMP4(path)
		if err != nil {
			r.log.Info("video: could not probe tracks, sending anyway",
				zap.String("strategy", label.strategy), zap.Error(err))
		} else if probe.Parsed && !probe.HasAudio() {
			r.log.Error("video: downloaded file has no audio track, refusing to send",
				zap.String("strategy", label.strategy), zap.String("detail", label.detail),
				zap.String("tracks", probe.String()), zap.Int64("bytes", fi.Size()),
				zap.String("hint", "ffmpeg is likely missing or not on PATH — "+
					"yt-dlp leaves video and audio unmerged and still reports success"))
			// Another strategy may still deliver both tracks in one file, so this
			// is a retry rather than a dead end.
			return outcomeRetry
		}
	}

	msg, err := sendVideo(ctx, r.bot, r.chatID, r.replyTo, path, metaFromInfo(info, caption))
	if err != nil {
		r.log.Info("video: send failed",
			zap.String("strategy", label.strategy), zap.String("detail", label.detail),
			zap.Error(err))
		return outcomeRetry
	}

	r.log.Info("video: sent successfully",
		zap.Int64("chat_id", r.chatID), zap.Int("message_id", msg.MessageID),
		zap.String("strategy", label.strategy), zap.String("detail", label.detail),
		zap.Int64("bytes", fi.Size()),
		zap.String("media", describeMediaFile(path)))

	if cacheVideo(r, storeKey, msg, info, title) {
		r.rememberAlias(aliasKey, storeKey)
	}
	return outcomeSent
}

// cacheVideo stores the uploaded video's file_id so the next time this link
// appears it is replayed from Telegram without any download. It reports whether
// anything was stored, so the caller knows whether an alias is worth recording.
func cacheVideo(r sendRequest, key string, msg *telego.Message, info *ytdlp.ExtractedInfo, title string) bool {
	if r.cache == nil || key == "" || msg == nil || msg.Video == nil {
		return false
	}

	entry := &media_cache.MediaCache{
		URL:      key,
		Kind:     media_cache.KindVideo,
		FileID:   msg.Video.FileID,
		Title:    title,
		Uploader: infoString(info.Uploader),
		Duration: msg.Video.Duration,
		Width:    msg.Video.Width,
		Height:   msg.Video.Height,
	}
	if err := r.cache.Put(entry); err != nil {
		r.log.Warn("video: failed to cache file_id",
			zap.String("key", key), zap.Error(err))
		return false
	}
	return true
}

// processPhotoPost scrapes a TikTok photo post and sends it as an album.
func processPhotoPost(ctx context.Context, r sendRequest, key string) bool {
	dir, err := os.MkdirTemp("", "famoria-photo-*")
	if err != nil {
		r.log.Error("photo: failed to create temp dir", zap.Error(err))
		return false
	}
	defer os.RemoveAll(dir)

	// The scraper needs the canonical post URL: a share link (vt./vm.) does not
	// carry the post ID that the page lookup is keyed on.
	fetchURL := r.fetchURL()

	// Photo posts honour the same proxy setting as yt-dlp when it applies to
	// all URLs; the youtube-only mode does not cover TikTok.
	proxySpec := ""
	if r.cfg.useProxy(fetchURL) {
		proxySpec = r.cfg.proxy
	}

	post, err := scrapeTikTokPhotoPost(ctx, fetchURL, dir, proxySpec)
	if err != nil {
		r.log.Info("photo: scrape failed",
			zap.String("url", fetchURL),
			zap.Bool("via_proxy", proxySpec != ""),
			zap.Error(err))
		return false
	}

	// The caption links back to the link the user pasted, not the resolved
	// canonical URL, so the attribution matches what they see in the chat.
	caption := buildCaption(r.lang, post.Title, r.originalURL)
	msgs, err := sendPhotoFiles(ctx, r, post.Images, caption)
	if err != nil {
		r.log.Info("photo: send failed",
			zap.String("url", fetchURL), zap.Error(err))
		return false
	}

	r.log.Info("photo: sent successfully",
		zap.Int64("chat_id", r.chatID), zap.Int("count", len(msgs)),
		zap.String("url", r.originalURL))

	cachePhotos(r, key, msgs, post.Title)
	return true
}

// cachePhotos stores the uploaded photo file_ids for replay.
func cachePhotos(r sendRequest, key string, msgs []telego.Message, title string) {
	if r.cache == nil || key == "" {
		return
	}

	ids := make([]string, 0, len(msgs))
	for _, m := range msgs {
		if id := largestPhotoFileID(m); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return
	}

	if err := r.cache.Put(&media_cache.MediaCache{
		URL:     key,
		Kind:    media_cache.KindPhoto,
		FileIDs: ids,
		Title:   title,
	}); err != nil {
		r.log.Warn("photo: failed to cache file_ids",
			zap.String("key", key), zap.Error(err))
	}
}

// largestPhotoFileID returns the file_id of the biggest size Telegram produced
// for a photo message. Telegram returns several resolutions and the last entry
// is the largest; replaying anything smaller would visibly degrade the cached
// send.
func largestPhotoFileID(m telego.Message) string {
	if len(m.Photo) == 0 {
		return ""
	}
	return m.Photo[len(m.Photo)-1].FileID
}

// sendPhotoFiles uploads local image files, then returns the sent messages.
func sendPhotoFiles(ctx context.Context, r sendRequest, paths []string, caption string) ([]telego.Message, error) {
	files := make([]*os.File, 0, len(paths))
	defer func() {
		for _, f := range files {
			_ = f.Close()
		}
	}()

	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	media := make([]telego.InputMedia, 0, len(files))
	for i, f := range files {
		// Each attachment needs a distinct name; telego uses it as the multipart
		// filename and as the attach:// reference.
		media = append(media, &telego.InputMediaPhoto{
			Type:  telego.MediaTypePhoto,
			Media: tu.FileFromReader(f, fmt.Sprintf("photo_%d.jpg", i+1)),
		})
	}
	return sendMedia(ctx, r, media, caption)
}

// sendPhotos replays cached photo file_ids.
func sendPhotos(ctx context.Context, r sendRequest, fileIDs []string, caption string) ([]telego.Message, error) {
	media := make([]telego.InputMedia, 0, len(fileIDs))
	for _, id := range fileIDs {
		media = append(media, &telego.InputMediaPhoto{
			Type:  telego.MediaTypePhoto,
			Media: tu.FileFromID(id),
		})
	}
	return sendMedia(ctx, r, media, caption)
}

// sendMedia delivers an album, falling back to a single photo send when the
// album holds one item.
//
// The split is required by the API: sendMediaGroup rejects fewer than 2 items,
// so a one-image post must go through sendPhoto instead.
func sendMedia(ctx context.Context, r sendRequest, media []telego.InputMedia, caption string) ([]telego.Message, error) {
	if len(media) == 0 {
		return nil, fmt.Errorf("no media to send")
	}

	reply := r.replyParams()

	// Only the first item of an album may carry the caption; Telegram renders it
	// as the caption of the whole group.
	if ph, ok := media[0].(*telego.InputMediaPhoto); ok {
		ph.Caption = caption
		ph.ParseMode = telego.ModeHTML
	}

	if len(media) == 1 {
		photo, ok := media[0].(*telego.InputMediaPhoto)
		if !ok {
			return nil, fmt.Errorf("single media item is not a photo")
		}
		msg, err := r.bot.SendPhoto(ctx, &telego.SendPhotoParams{
			ChatID:          tu.ID(r.chatID),
			Photo:           photo.Media,
			Caption:         caption,
			ParseMode:       telego.ModeHTML,
			ReplyParameters: reply,
		})
		if err != nil {
			return nil, err
		}
		return []telego.Message{*msg}, nil
	}

	return r.bot.SendMediaGroup(ctx, &telego.SendMediaGroupParams{
		ChatID:          tu.ID(r.chatID),
		Media:           media,
		ReplyParameters: reply,
	})
}

// replyParams builds the reply target, or nil when the message being answered
// no longer exists. AllowSendingWithoutReply keeps the media from failing just
// because the source message was deleted meanwhile.
func (r sendRequest) replyParams() *telego.ReplyParameters {
	if r.replyTo == 0 {
		return nil
	}
	return &telego.ReplyParameters{
		MessageID:                r.replyTo,
		AllowSendingWithoutReply: true,
	}
}
