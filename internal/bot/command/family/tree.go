package family

import (
	"bytes"
	"context"
	"errors"
	"famoria/internal/config"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type treeCmd struct {
	cfg config.Config
	log *zap.Logger
}

type Image int

func (i Image) String() string {
	switch i {
	case Text:
		return "text"
	case Graphviz:
		return "image_graphviz"
	case Ete3:
		return "image_ete3"
	case AnyTree:
		return "image_igraph"
	case Networkx:
		return "image_networkx"
	default:
		return "text"
	}
}

const (
	Text Image = iota
	Graphviz
	Ete3
	AnyTree
	Networkx
)

func (c treeCmd) Handle(ctx *th.Context, update telego.Update) error {
	args := strings.Split(update.Message.Text, " ")
	mode := "text"
	if len(args) > 1 {
		arg, err := strconv.Atoi(args[1])
		if err == nil {
			mode = Image(arg).String()
		}
	}
	requestURL := fmt.Sprintf("%s/tree/%s/%d?reverse=true", c.cfg.TreeApiURL, mode, update.Message.From.ID)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		c.log.Sugar().Error("client: could not create request: %s\n", err)
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.log.Sugar().Error("client: error making http request: %s\n", err)
		return err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		c.log.Sugar().Error("client: could not read response body: %s\n", err)
		return err
	}

	switch mode {
	case "text":
		_, err = ctx.Bot().SendMessage(
			context.Background(),
			&telego.SendMessageParams{
				ChatID: tu.ID(update.Message.Chat.ID),
				Text:   string(resBody),
			},
		)
	default:
		contentType := res.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			c.log.Sugar().Error("client: expected image but got: %s", contentType)
			return errors.New("expected image but got: " + contentType)
		}

		_, err = ctx.Bot().SendPhoto(context.Background(), &telego.SendPhotoParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Photo: tu.File(
				tu.NameReader(
					bytes.NewReader(resBody),
					"treeCmd.png",
				),
			),
		})
	}

	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
