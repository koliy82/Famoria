package buttons

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type Builder struct {
	MaxRows int
	MaxCols int
	Buttons [][]telego.InlineKeyboardButton
}

func New(maxRows int, maxCols int) *Builder {
	return &Builder{
		MaxRows: maxRows,
		MaxCols: maxCols,
		Buttons: make([][]telego.InlineKeyboardButton, 0),
	}
}

func (b *Builder) Add(button telego.InlineKeyboardButton) {
	if len(b.Buttons) == 0 || len(b.Buttons[len(b.Buttons)-1]) == b.MaxCols {
		b.Buttons = append(b.Buttons, make([]telego.InlineKeyboardButton, 0))
	}
	b.Buttons[len(b.Buttons)-1] = append(b.Buttons[len(b.Buttons)-1], button)
}

func (b *Builder) Build() telego.ReplyMarkup {
	return tu.InlineKeyboardGrid(b.Buttons)
}

func (b *Builder) InlineBuild() *telego.InlineKeyboardMarkup {
	return tu.InlineKeyboardGrid(b.Buttons)
}
