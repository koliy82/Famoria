package family

import (
	"bytes"
	"famoria/internal/config"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func (c treeCmd) Handle(bot *telego.Bot, update telego.Update) {
	args := strings.Split(update.Message.Text, " ")
	mode := "text"
	if len(args) > 1 {
		arg, err := strconv.Atoi(args[1])
		if err == nil {
			mode = Image(arg).String()
		}
	}
	requestURL := fmt.Sprintf("%s/treeCmd/%s/%d?reverse=true", c.cfg.ApiURL, mode, update.Message.From.ID)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		c.log.Sugar().Error("client: could not create request: %s\n", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.log.Sugar().Error("client: error making http request: %s\n", err)
		return
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		c.log.Sugar().Error("client: could not read response body: %s\n", err)
		return
	}

	switch mode {
	case "text":
		_, err = bot.SendMessage(
			&telego.SendMessageParams{
				ChatID: tu.ID(update.Message.Chat.ID),
				Text:   string(resBody),
			},
		)
	default:
		contentType := res.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			c.log.Sugar().Error("client: expected image but got: %s", contentType)
			return
		}

		_, err = bot.SendPhoto(&telego.SendPhotoParams{
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
}
