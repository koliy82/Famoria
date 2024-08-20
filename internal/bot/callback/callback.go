package callback

import (
	"github.com/koliy82/telego"
	th "github.com/koliy82/telego/telegohandler"
	tu "github.com/koliy82/telego/telegoutil"
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

func (callback *Callback) Inline() telego.InlineKeyboardButton {
	return tu.InlineKeyboardButton(callback.Label).
		WithCallbackData(callback.Data)
}

func Register(bh *th.BotHandler, cm *CallbacksManager) {
	bh.HandleCallbackQuery(
		func(bot *telego.Bot, query telego.CallbackQuery) {
			cm.HandleCallback(bot, query)
		},
		th.AnyCallbackQuery(),
	)
}
