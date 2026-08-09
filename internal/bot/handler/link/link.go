package link

import (
	"context"
	"encoding/json"
	"famoria/internal/pkg/common/extractor"
	"os"

	"github.com/lrstanley/go-ytdlp"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type AnyLinkDownloader struct {
	log  *zap.Logger
	ytdl *ytdlp.Command
}

func (l AnyLinkDownloader) Handle(ctx *th.Context, update telego.Update) error {
	params := &telego.SendMessageParams{
		ChatID: tu.ID(update.Message.Chat.ID),
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.GetMessageID(),
			AllowSendingWithoutReply: true,
		},
		DisableNotification: true,
	}
	var userUrl, userUrlErr = extractor.ExtractLink(update.Message.Text)
	if userUrlErr != nil {
		_, err := ctx.Bot().SendMessage(context.Background(), params.WithText("failed to extract link"))
		l.log.Sugar().Error("failed to extract link: "+userUrl.String(), zap.Error(userUrlErr))
		return err
	}
	l.log.Sugar().Info("extracted link: " + userUrl.String())
	r, err := l.ytdl.Run(context.TODO(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err != nil {
		panic(err)
	}
	f, err := os.Create("results.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")

	if err = enc.Encode(r); err != nil {
		panic(err)
	}
	l.log.Info("wrote results to results.json")
	return nil
}

type Opts struct {
	fx.In
	Bh  *th.BotHandler
	Log *zap.Logger
}

func Register(opts Opts) {
	ytdlp.MustInstall(context.TODO(), nil)
	opts.Bh.Handle(AnyLinkDownloader{
		log: opts.Log,
		ytdl: ytdlp.New().
			FormatSort("res,ext:mp4:m4a").
			RecodeVideo("mp4").
			Output("%(extractor)s - %(title)s.%(ext)s"),
	}.Handle, th.TextMatches(extractor.LinkRegex))
}
