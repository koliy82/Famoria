package link

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"famoria/internal/pkg/html"

	"github.com/lrstanley/go-ytdlp"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

const (
	maxDurationSeconds = 30 * 60      // skip videos longer than 30 minutes
	maxFileBytes       = 50 * 1 << 20 // Telegram bot upload limit: 50 MB
	perVideoTimeout    = 10 * 60      // 10 minutes per video processing
)

// qualityCascade is tried in order until the downloaded file fits under the
// 50 MB upload limit.
var qualityCascade = []int{720, 480, 144}

// isYouTube reports whether the URL points at a YouTube host. yt-dlp needs
// cookies for YouTube because it blocks unauthenticated (bot) access.
func isYouTube(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Host)
	return host == "youtube.com" || host == "www.youtube.com" ||
		host == "m.youtube.com" || host == "music.youtube.com" ||
		host == "youtu.be" || strings.HasSuffix(host, ".youtube.com")
}

// applyCookies attaches a cookies file to the yt-dlp command when the URL is a
// YouTube link and a cookies file is configured. Returns the command unchanged
// otherwise.
func applyCookies(cmd *ytdlp.Command, rawURL, cookiesFile string) *ytdlp.Command {
	if cookiesFile == "" || !isYouTube(rawURL) {
		return cmd
	}
	return cmd.Cookies(cookiesFile)
}

// extractInfo queries yt-dlp for metadata about the URL without downloading the
// media. Returns the first extracted info entry, or an error if the URL is not
// a supported video.
func extractInfo(ctx context.Context, rawURL, cookiesFile string) (*ytdlp.ExtractedInfo, error) {
	cmd := ytdlp.New().
		DumpJSON().
		Simulate().
		NoPlaylist().
		NoWarnings().
		Quiet().
		NoColors()
	cmd = applyCookies(cmd, rawURL, cookiesFile)

	r, err := cmd.Run(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	info, err := r.GetExtractedInfo()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, fmt.Errorf("no extractable video info for %s", rawURL)
	}
	return info[0], nil
}

// downloadVideo downloads a single video at the requested target height into
// dir and returns the path to the resulting mp4 file.
//
// Telegram only plays H.264 video + AAC audio; other codecs (h265/HEVC, VP9,
// Opus) show up as silent video. The format selection therefore prefers
// H.264+AAC streams while keeping the audio track intact:
//
//  1. A pre-muxed H.264+AAC stream — TikTok / Reels ship only combined
//     streams, and crucially their H.264 variant may exist only at a lower
//     resolution (e.g. 540p) while 720p+ is HEVC. Filtering by [height<=N]
//     here would discard the only audible format, so resolution is steered
//     via FormatSort (a preference) instead of a hard filter.
//  2. Separate best H.264 video + AAC audio for DASH sources (YouTube),
//     remuxed (not re-encoded) via MergeOutputFormat.
//  3. Generic best fallback.
//
// RecodeVideo must NOT be used: it re-encodes the whole file and was reliably
// stripping the audio on already-combined streams (e.g. TikTok).
func downloadVideo(ctx context.Context, rawURL, dir string, height int, cookiesFile string) (string, error) {
	cmd := ytdlp.New().
		Format("best[vcodec^=h264][acodec^=aac]/" +
			"bestvideo[vcodec^=h264]+bestaudio[acodec^=aac]/" +
			"best[acodec^=aac]/best").
		// res:N is a preference (not a filter): pick the format closest to the
		// target height, then prefer h264/aac/mp4 for Telegram compatibility.
		FormatSort(fmt.Sprintf("res:%d,vcodec:h264,acodec:aac,ext:mp4:m4a", height)).
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
	cmd = applyCookies(cmd, rawURL, cookiesFile)

	if _, err := cmd.Run(ctx, rawURL); err != nil {
		return "", err
	}
	return findMP4(dir)
}

// findMP4 returns the path to the largest .mp4 file in dir, or an error if none
// exists. yt-dlp may leave intermediate files, so picking the largest reliably
// identifies the final merged output.
func findMP4(dir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.mp4"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no mp4 file produced in %s", dir)
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
	return best, nil
}

// sendVideo uploads the video file to the chat with the standard caption
// linking back to the original URL. The file is uploaded with a clean
// "video.mp4" name and the metadata (duration/width/height) from yt-dlp, which
// lets Telegram process the audio track correctly. Without an explicit
// .mp4 filename Telegram may misclassify the upload and drop the sound.
func sendVideo(ctx context.Context, bot *telego.Bot, chatID int64, replyTo int, path, originalURL string, info *ytdlp.ExtractedInfo) (*telego.Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	params := &telego.SendVideoParams{
		ChatID:            tu.ID(chatID),
		Video:             tu.FileFromReader(f, "video.mp4"),
		Caption:           html.Link(originalURL, "Оригинал") + " | converted by Famoria",
		ParseMode:         telego.ModeHTML,
		SupportsStreaming: true,
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                replyTo,
			AllowSendingWithoutReply: true,
		},
	}
	if info.Duration != nil {
		params.Duration = int(*info.Duration)
	}
	if info.Width != nil {
		params.Width = int(*info.Width)
	}
	if info.Height != nil {
		params.Height = int(*info.Height)
	}
	return bot.SendVideo(ctx, params)
}

// processVideo runs the full pipeline for one link: metadata check, quality
// cascade download, and send. Returns true if the video was sent successfully
// (so the caller can delete the original message).
func processVideo(ctx context.Context, bot *telego.Bot, log *zap.Logger, chatID int64, replyTo int, originalURL, cookiesFile string) bool {
	info, err := extractInfo(ctx, originalURL, cookiesFile)
	if err != nil {
		log.Info("video: not a supported video or extract failed",
			zap.String("url", originalURL), zap.Error(err))
		return false
	}

	if info.Duration != nil && *info.Duration > maxDurationSeconds {
		log.Info("video: skipping, too long",
			zap.String("url", originalURL), zap.Float64("duration", *info.Duration))
		return false
	}

	dir, err := os.MkdirTemp("", "famoria-video-*")
	if err != nil {
		log.Sugar().Error("video: failed to create temp dir", zap.Error(err))
		return false
	}
	defer os.RemoveAll(dir)

	for _, height := range qualityCascade {
		path, err := downloadVideo(ctx, originalURL, dir, height, cookiesFile)
		if err != nil {
			log.Info("video: download failed at resolution",
				zap.Int("height", height), zap.String("url", originalURL), zap.Error(err))
			continue
		}

		fi, err := os.Stat(path)
		if err != nil {
			log.Info("video: could not stat downloaded file",
				zap.Int("height", height), zap.Error(err))
			continue
		}
		if fi.Size() >= maxFileBytes {
			log.Info("video: file too large, trying lower resolution",
				zap.Int("height", height), zap.Int64("bytes", fi.Size()))
			_ = os.Remove(path)
			continue
		}

		msg, err := sendVideo(ctx, bot, chatID, replyTo, path, originalURL, info)
		if err != nil {
			log.Info("video: send failed",
				zap.Int("height", height), zap.String("url", originalURL), zap.Error(err))
			_ = os.Remove(path)
			continue
		}
		log.Info("video: sent successfully",
			zap.Int64("chat_id", chatID), zap.Int("message_id", msg.MessageID),
			zap.Int("height", height), zap.Int64("bytes", fi.Size()))
		return true
	}

	log.Info("video: could not fit under 50MB at any resolution",
		zap.String("url", originalURL))
	return false
}
