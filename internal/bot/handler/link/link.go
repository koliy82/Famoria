package link

import (
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/config"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/pkg/common/extractor"
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
	cookiesFile  string
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

	sent := processVideo(ctx, l.bot, l.log, chatID, msg.MessageID, originalURL, l.cookiesFile)
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

	var cookiesFile string
	if opts.Cfg.YtdlpCookiesFile != nil {
		cookiesFile = *opts.Cfg.YtdlpCookiesFile
	}

	dl := AnyLinkDownloader{
		log:          opts.Log,
		bot:          opts.Bot,
		chatSettings: opts.ChatSettings,
		cookiesFile:  cookiesFile,
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
