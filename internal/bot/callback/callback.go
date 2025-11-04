package callback

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
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
		func(ctx *th.Context, query telego.CallbackQuery) error {
			cm.HandleCallback(ctx, query)
			return nil
		},
		th.AnyCallbackQuery(),
	)
}
