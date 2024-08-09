package family

import (
	"famoria/internal/config"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type tree struct {
	cfg  config.Config
	log  *zap.Logger
	mode string
}

func (t tree) Handle(bot *telego.Bot, update telego.Update) {
	requestURL := fmt.Sprintf("%s/tree/%s/%d?reverse=true", t.cfg.ApiURL, t.mode, update.Message.From.ID)
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
	t.log.Sugar().Debug(fmt.Sprintf("client: response body: %s\n", resBody))

	switch t.mode {
	case "text":
		_, err = bot.SendMessage(
			&telego.SendMessageParams{
				ChatID:    tu.ID(update.Message.Chat.ID),
				ParseMode: telego.ModeHTML,
				Text:      string(resBody),
			},
		)
	case "image":
		_, err = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"%s, данная команда находится в разработке..", update.Message.From.FirstName,
		))
		//_, err = bot.SendPhoto(&telego.SendPhotoParams{
		//	ChatID: tu.ID(update.Message.Chat.ID),
		//})
	}

	if err != nil {
		t.log.Sugar().Error(err)
	}
}
