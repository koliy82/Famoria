package callback

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"
)

type ContextType string

const (
	Static    ContextType = "STATIC"
	Temporary ContextType = "TEMPORARY"
	OneClick  ContextType = "ONE_CLICK"
	ChooseOne ContextType = "CHOOSE_ONE"
)

type Callback struct {
	Data       string
	Type       ContextType
	OwnerIDs   []int64
	Label      string
	AnswerText string
	Callback   func(query telego.CallbackQuery)
}

func Register(bh *th.BotHandler, log *zap.Logger, cm *CallbacksManager) {
	bh.HandleCallbackQuery(
		func(bot *telego.Bot, query telego.CallbackQuery) {
			cm.HandleCallback(bot, query, log)
		},
		th.AnyCallbackQuery(),
	)
}
