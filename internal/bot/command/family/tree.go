package family

import (
	"bytes"
	"famoria/internal/config"
	"fmt"
	"github.com/koliy82/telego"
	tu "github.com/koliy82/telego/telegoutil"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type tree struct {
	cfg config.Config
	log *zap.Logger
}

type Image int

func (i Image) String() string {
	switch i {
	case Graphviz:
		return "graphviz"
	case Ete3:
		return "ete3"
	case AnyTree:
		return "anytree"
	case Igraph:
		return "igraph"
	case Plotly:
		return "v2"
	case Networkx:
		return "v3"
	default:
		return "graphviz"
	}
}

const (
	Graphviz Image = iota
	Ete3
	AnyTree
	Igraph
	Plotly
	Networkx
)

func (t tree) Handle(bot *telego.Bot, update telego.Update) {
	args := strings.Split(update.Message.Text, " ")
	mode := "text"
	if len(args) > 1 {
		arg, err := strconv.Atoi(args[1])
		if err == nil {
			mode = "image_" + Image(arg).String()
		}
	}
	requestURL := fmt.Sprintf("%s/tree/%s/%d?reverse=true", t.cfg.ApiURL, mode, update.Message.From.ID)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		t.log.Sugar().Error("client: could not create request: %s\n", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.log.Sugar().Error("client: error making http request: %s\n", err)
		return
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.log.Sugar().Error("client: could not read response body: %s\n", err)
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
			t.log.Sugar().Error("client: expected image but got: %s", contentType)
			return
		}

		_, err = bot.SendPhoto(&telego.SendPhotoParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Photo: tu.File(
				tu.NameReader(
					bytes.NewReader(resBody),
					"tree.png",
				),
			),
		})
	}

	if err != nil {
		t.log.Sugar().Error(err)
	}
}
