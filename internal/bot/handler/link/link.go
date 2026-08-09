package link

import (
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/config"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/pkg/common/extractor"
	"os"
	"strings"
	"time"

	"github.com/lrstanley/go-ytdlp"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type AnyLinkDownloader struct {
	log          *zap.Logger
	bot          *telego.Bot
	chatSettings chat_settings.Repository
	ytcfg        ytConfig
	queue        *queueManager
}

func (l AnyLinkDownloader) Handle(ctx *th.Context, update telego.Update) error {
	msg := update.Message
	if msg == nil || msg.From == nil {
		return nil
	}

	// Only process in chats where the converter is enabled.
	if !l.chatSettings.IsVideoConverterEnabled(msg.Chat.ID) {
		return nil
	}

	userURL, err := extractor.ExtractLink(msg.Text)
	if err != nil {
		l.log.Info("link: failed to extract link", zap.String("text", msg.Text), zap.Error(err))
		return nil
	}
	l.log.Info("link: extracted", zap.String("url", userURL.String()))

	// Enqueue and return immediately; the worker processes the job off the
	// handler goroutine so long polling is never blocked.
	l.queue.submit(msg.Chat.ID, job{
		update: update,
		url:    userURL,
	})
	return nil
}

// processJob is the per-chat worker callback. It performs the heavy work:
// metadata check, download cascade, send, and deletion of the original message.
func (l AnyLinkDownloader) processJob(j job) {
	msg := j.update.Message
	chatID := msg.Chat.ID
	originalURL := j.url.String()

	ctx, cancel := context.WithTimeout(context.Background(), perVideoTimeout*time.Second)
	defer cancel()

	sent := processVideo(ctx, l.bot, l.log, chatID, msg.MessageID, originalURL, l.ytcfg)
	if !sent {
		return
	}

	// Video was sent successfully — remove the original link message.
	if err := l.bot.DeleteMessage(context.Background(), &telego.DeleteMessageParams{
		ChatID:    tu.ID(chatID),
		MessageID: msg.MessageID,
	}); err != nil {
		l.log.Info("link: failed to delete original message",
			zap.Int64("chat_id", chatID), zap.Int("message_id", msg.MessageID), zap.Error(err))
	}
}

type Opts struct {
	fx.In
	Bh           *th.BotHandler
	Log          *zap.Logger
	Bot          *telego.Bot
	Cm           *callback.CallbacksManager
	Cfg          config.Config
	ChatSettings chat_settings.Repository
}

func Register(opts Opts) {
	// Ensure yt-dlp is available. Prefer a system-installed yt-dlp (e.g. via
	// apk/yum in the Docker image) at any version; only download from GitHub
	// as a fallback (local dev without yt-dlp on PATH). This avoids re-downloading
	// ~30MB on every container start and survives environments without outbound
	// network at runtime.
	if _, err := ytdlp.Install(context.TODO(), &ytdlp.InstallOptions{
		AllowVersionMismatch: true,
	}); err != nil {
		opts.Log.Warn("ytdlp: system binary not found, downloading", zap.Error(err))
		ytdlp.MustInstall(context.TODO(), nil)
	}

	var ytcfg ytConfig
	if opts.Cfg.YtdlpCookiesFile != nil {
		ytcfg.cookiesFile = *opts.Cfg.YtdlpCookiesFile
	}
	if opts.Cfg.ProxyURL != nil {
		ytcfg.proxy = *opts.Cfg.ProxyURL
	}
	// Normalize the proxy mode; anything outside the known values means "off".
	switch strings.ToLower(strings.TrimSpace(opts.Cfg.ProxyEnable)) {
	case proxyModeAll, "1":
		ytcfg.proxyMode = proxyModeAll
	case proxyModeYouTube, "yt":
		ytcfg.proxyMode = proxyModeYouTube
	default:
		ytcfg.proxyMode = proxyModeFalse
	}

	// Log the cookies configuration so misconfigurations (wrong path, missing
	// file) are visible at startup rather than surfacing as opaque "Sign in to
	// confirm you're not a bot" errors at runtime.
	if ytcfg.cookiesFile == "" {
		opts.Log.Warn("ytdlp: no cookies file configured (YTDLP_COOKIES_FILE empty) — YouTube will block downloads")
	} else {
		if _, err := os.Stat(ytcfg.cookiesFile); err != nil {
			opts.Log.Error("ytdlp: cookies file not accessible — YouTube will block downloads",
				zap.String("path", ytcfg.cookiesFile), zap.Error(err))
		} else {
			opts.Log.Info("ytdlp: cookies file loaded", zap.String("path", ytcfg.cookiesFile))
		}
	}

	// Log the proxy configuration. Mask credentials in the URL so the password
	// isn't leaked into shared logs.
	if ytcfg.proxy == "" {
		opts.Log.Info("ytdlp: proxy disabled (PROXY_URL empty)")
	} else {
		opts.Log.Info("ytdlp: proxy configured",
			zap.String("mode", ytcfg.proxyMode), zap.String("proxy", maskProxyCreds(ytcfg.proxy)))
	}

	dl := AnyLinkDownloader{
		log:          opts.Log,
		bot:          opts.Bot,
		chatSettings: opts.ChatSettings,
		ytcfg:        ytcfg,
	}
	dl.queue = newQueueManager(dl.processJob)

	// /settings command for chat admins.
	opts.Bh.Handle(settingsCmd{
		bot:          opts.Bot,
		log:          opts.Log,
		chatSettings: opts.ChatSettings,
		cm:           opts.Cm,
	}.Handle, th.CommandEqual("settings"))

	// Link handler — matches any text containing a link.
	opts.Bh.Handle(dl.Handle, th.TextMatches(extractor.LinkRegex))
}
