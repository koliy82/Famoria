package showInventory

import (
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/inventory"
	"famoria/internal/bot/idle/item/items"
	"famoria/internal/database/mongo/repositories/brak"
	"fmt"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type Inventory struct {
	Items         []*inventory.ShowItem
	ShowCallbacks [][]telego.InlineKeyboardButton
	NavigateBack  callback.Callback
	NavigateNext  callback.Callback
	MaxRows       int
	MaxCells      int
	MaxPages      int
	CurrentPage   int
	Label         string
	SelectedItem  *inventory.ShowItem
	Opts          *Opts
}

type Opts struct {
	B              *brak.Brak
	Params         *telego.SendMessageParams
	BotCtx         *th.Context
	Manager        *item.Manager
	Log            *zap.Logger
	Cm             *callback.CallbacksManager
	InventoryItems map[items.Name]inventory.Item
}

func New(opts *Opts) *Inventory {
	p := &Inventory{
		Items:       make([]*inventory.ShowItem, 0, len(opts.InventoryItems)),
		CurrentPage: 1,
		MaxRows:     3,
		MaxCells:    3,
		Opts:        opts,
	}
	for _, i := range opts.InventoryItems {
		mi := opts.Manager.Items[i.Name]
		println(i.Name.String())
		println(i.CurrentLevel)
		p.Items = append(p.Items, &inventory.ShowItem{
			Emoji:        mi.Emoji,
			Name:         mi.Name,
			CurrentLevel: i.CurrentLevel,
			MaxLevel:     mi.MaxLevel,
			Description:  mi.Description,
			Buffs:        mi.Buffs[i.CurrentLevel],
		})
	}
	if len(p.Items) == 0 {
		_, err := opts.BotCtx.Bot().SendMessage(context.Background(), opts.Params.WithText("Инвентарь пуст, милорд."))
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
		return nil
	}
	p.MaxPages = len(p.Items) / (p.MaxRows * p.MaxCells)
	if len(p.Items)%(p.MaxRows*p.MaxCells) != 0 {
		p.MaxPages++
	}
	if p.MaxPages > 1 {
		p.SetNavigateButtons()
	}
	p.CurrentButtonsPage()
	return p
}

func (p *Inventory) CurrentButtonsPage() [][]telego.InlineKeyboardButton {
	p.Label = fmt.Sprintf("Инвентарь (%d/%d стр.)\n", p.CurrentPage, p.MaxPages)
	for i := 0; i < len(p.ShowCallbacks)-1; i++ {
		for j := 0; j < len(p.ShowCallbacks[i]); j++ {
			p.Opts.Cm.RemoveCallback(p.ShowCallbacks[i][j].CallbackData)
		}
	}
	p.ShowCallbacks = make([][]telego.InlineKeyboardButton, 0, p.MaxRows)
	for i := 0; i < p.MaxRows; i++ {
		if i*p.MaxCells >= len(p.Items) {
			break
		}
		row := make([]telego.InlineKeyboardButton, 0, p.MaxCells)
		for j := 0; j < p.MaxCells; j++ {
			if i*p.MaxCells+j >= len(p.Items) {
				break
			}
			si := p.Items[i*p.MaxCells+j]
			//p.Label += si.SmallDescription() + "\n"
			iCallback := p.Opts.Cm.DynamicCallback(callback.DynamicOpts{
				Label:    si.Name.String(),
				CtxType:  callback.Temporary,
				OwnerIDs: []int64{p.Opts.B.FirstUserID, p.Opts.B.SecondUserID},
				Time:     time.Duration(30) * time.Minute,
				Callback: func(query telego.CallbackQuery) {
					if p.SelectedItem != nil && p.SelectedItem.Name == si.Name {
						_ = p.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
							CallbackQueryID: query.ID,
						})
						return
					}
					p.SelectedItem = si
					_, err := p.Opts.BotCtx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
						MessageID: query.Message.GetMessageID(),
						ChatID:    tu.ID(query.Message.GetChat().ID),
						ParseMode: telego.ModeHTML,
						Text:      p.Label + si.FullDescription(),
						ReplyMarkup: &telego.InlineKeyboardMarkup{
							InlineKeyboard: p.ShowCallbacks,
						},
					})
					if err != nil {
						p.Opts.Log.Sugar().Error(err)
					}
					_ = p.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(),
						&telego.AnswerCallbackQueryParams{
							CallbackQueryID: query.ID,
						})
				},
			})
			row = append(row, iCallback.Inline())
		}
		p.ShowCallbacks = append(p.ShowCallbacks, row)
	}
	if p.MaxPages > 1 {
		p.ShowCallbacks = append(p.ShowCallbacks, []telego.InlineKeyboardButton{p.NavigateBack.Inline(), p.NavigateNext.Inline()})
	}
	return p.ShowCallbacks
}

func (p *Inventory) SetNavigateButtons() {
	p.NavigateBack = p.Opts.Cm.DynamicCallback(callback.DynamicOpts{
		Label:    "⬅️",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{p.Opts.B.FirstUserID, p.Opts.B.SecondUserID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, err := p.Opts.BotCtx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
				MessageID: query.Message.GetMessageID(),
				ChatID:    tu.ID(query.Message.GetChat().ID),
				ParseMode: telego.ModeHTML,
				Text:      p.Label + "Выберите предмет для просмотра.",
				ReplyMarkup: &telego.InlineKeyboardMarkup{
					InlineKeyboard: p.PrevPageButtons(),
				},
			})
			if err != nil {
				p.Opts.Log.Sugar().Error(err)
			}
			_ = p.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
			})
		}})

	p.NavigateNext = p.Opts.Cm.DynamicCallback(callback.DynamicOpts{
		Label:    "➡️",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{p.Opts.B.FirstUserID, p.Opts.B.SecondUserID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, err := p.Opts.BotCtx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
				MessageID: query.Message.GetMessageID(),
				ChatID:    tu.ID(query.Message.GetChat().ID),
				ParseMode: telego.ModeHTML,
				Text:      p.Label + "Выберите предмет для просмотра.",
				ReplyMarkup: &telego.InlineKeyboardMarkup{
					InlineKeyboard: p.NextPageButtons(),
				},
			})
			if err != nil {
				p.Opts.Log.Sugar().Error(err)
			}
			_ = p.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
			})
		}})
}

func (p *Inventory) NextPageButtons() [][]telego.InlineKeyboardButton {
	if p.CurrentPage == p.MaxPages {
		p.CurrentPage = 1
	} else {
		p.CurrentPage++
	}
	return p.CurrentButtonsPage()
}

func (p *Inventory) PrevPageButtons() [][]telego.InlineKeyboardButton {
	if p.CurrentPage == 1 {
		p.CurrentPage = p.MaxPages
	} else {
		p.CurrentPage--
	}
	return p.CurrentButtonsPage()
}
