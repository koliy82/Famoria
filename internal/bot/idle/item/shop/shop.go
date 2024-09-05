package shop

import (
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/inventory"
	"famoria/internal/database/mongo/repositories/brak"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type Shop struct {
	Items         []*inventory.ShopItem
	ShopCallbacks [][]telego.InlineKeyboardButton
	BackCallback  callback.Callback
	BuyCallback   callback.Callback
	NavigateBack  callback.Callback
	NavigateNext  callback.Callback
	MaxRows       int
	MaxCells      int
	MaxPages      int
	CurrentPage   int
	Label         string
	SelectedItem  *inventory.ShopItem
	Opts          *Opts
}

type Opts struct {
	B       *brak.Brak
	Params  *telego.SendMessageParams
	Bot     *telego.Bot
	Manager *item.Manager
	Log     *zap.Logger
	Cm      *callback.CallbacksManager
}

func New(opts *Opts) *Shop {
	s := &Shop{
		CurrentPage: 1,
		MaxRows:     3,
		MaxCells:    4,
		Opts:        opts,
	}
	s.Items = opts.B.Inventory.GetAvailableItems(opts.Manager)
	if len(s.Items) == 0 {
		_, err := opts.Bot.SendMessage(opts.Params.
			WithText("Вы скупили все доступные предметы на данный момент, милорд."),
		)
		if err != nil {
			opts.Log.Sugar().Error(err)
		}
		return nil
	}
	//for i := 0; i < 4; i++ {
	//	s.Items = append(s.Items, s.Items...)
	//}
	s.MaxPages = len(s.Items) / (s.MaxRows * s.MaxCells)
	if len(s.Items)%(s.MaxRows*s.MaxCells) != 0 {
		s.MaxPages++
	}
	s.SetItemCallbacks()
	if s.MaxPages > 1 {
		s.SetNavigateButtons()
	}
	s.CurrentButtonsPage()
	return s
}

func (s *Shop) CurrentButtonsPage() [][]telego.InlineKeyboardButton {
	s.Label = fmt.Sprintf("Потайная лавка (%d/%d)\n", s.CurrentPage, s.MaxPages)
	for i := 0; i < len(s.ShopCallbacks)-1; i++ {
		for j := 0; j < len(s.ShopCallbacks[i]); j++ {
			s.Opts.Cm.RemoveCallback(s.ShopCallbacks[i][j].CallbackData)
		}
	}
	s.ShopCallbacks = make([][]telego.InlineKeyboardButton, 0)
	startIndex := (s.CurrentPage - 1) * s.MaxRows * s.MaxCells
	endIndex := startIndex + s.MaxRows*s.MaxCells
	if endIndex > len(s.Items) {
		endIndex = len(s.Items)
	}
	for i := 0; i < s.MaxRows; i++ {
		s.ShopCallbacks = append(s.ShopCallbacks, make([]telego.InlineKeyboardButton, 0))
		for j := 0; j < s.MaxCells; j++ {
			itemIndex := startIndex + i*s.MaxCells + j
			if itemIndex >= endIndex {
				break
			}
			si := s.Items[itemIndex]
			s.Label += si.SmallDescription() + "\n"
			dCallback := s.Opts.Cm.DynamicCallback(callback.DynamicOpts{
				Label:    fmt.Sprintf("%s %d/%d", si.Emoji, si.BuyLevel, si.MaxLevel),
				CtxType:  callback.Temporary,
				OwnerIDs: []int64{s.Opts.B.FirstUserID, s.Opts.B.SecondUserID},
				Time:     time.Duration(30) * time.Minute,
				Callback: func(query telego.CallbackQuery) {
					s.SelectedItem = si
					_, err := s.Opts.Bot.EditMessageText(&telego.EditMessageTextParams{
						MessageID: query.Message.GetMessageID(),
						ChatID:    tu.ID(query.Message.GetChat().ID),
						ParseMode: telego.ModeHTML,
						Text:      si.FullDescription(),
						ReplyMarkup: tu.InlineKeyboard(
							tu.InlineKeyboardRow(
								s.BackCallback.Inline(),
								s.BuyCallback.Inline(),
							),
						),
					})
					if err != nil {
						s.Opts.Log.Sugar().Error(err)
					}
				},
			})
			s.ShopCallbacks[i] = append(s.ShopCallbacks[i], dCallback.Inline())
		}
	}

	if s.MaxPages > 1 {
		s.ShopCallbacks = append(s.ShopCallbacks, []telego.InlineKeyboardButton{s.NavigateBack.Inline(), s.NavigateNext.Inline()})
	}

	return s.ShopCallbacks
}

func (s *Shop) NextPageButtons() [][]telego.InlineKeyboardButton {
	if s.CurrentPage == s.MaxPages {
		s.CurrentPage = 1
	} else {
		s.CurrentPage++
	}
	return s.CurrentButtonsPage()
}

func (s *Shop) PrevPageButtons() [][]telego.InlineKeyboardButton {
	if s.CurrentPage == 1 {
		s.CurrentPage = s.MaxPages
	} else {
		s.CurrentPage--
	}
	return s.CurrentButtonsPage()
}

func (s *Shop) SetNavigateButtons() {
	s.NavigateBack = s.Opts.Cm.DynamicCallback(callback.DynamicOpts{
		Label:    "⬅️",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{s.Opts.B.FirstUserID, s.Opts.B.SecondUserID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, err := s.Opts.Bot.EditMessageText(&telego.EditMessageTextParams{
				MessageID: query.Message.GetMessageID(),
				ChatID:    tu.ID(query.Message.GetChat().ID),
				ParseMode: telego.ModeHTML,
				Text:      s.Label,
				ReplyMarkup: &telego.InlineKeyboardMarkup{
					InlineKeyboard: s.PrevPageButtons(),
				},
			})
			if err != nil {
				s.Opts.Log.Sugar().Error(err)
			}
		}})

	s.NavigateNext = s.Opts.Cm.DynamicCallback(callback.DynamicOpts{
		Label:    "➡️",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{s.Opts.B.FirstUserID, s.Opts.B.SecondUserID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, err := s.Opts.Bot.EditMessageText(&telego.EditMessageTextParams{
				MessageID: query.Message.GetMessageID(),
				ChatID:    tu.ID(query.Message.GetChat().ID),
				ParseMode: telego.ModeHTML,
				Text:      s.Label,
				ReplyMarkup: &telego.InlineKeyboardMarkup{
					InlineKeyboard: s.NextPageButtons(),
				},
			})
			if err != nil {
				s.Opts.Log.Sugar().Error(err)
			}
		}})
}

func (s *Shop) SetItemCallbacks() {
	s.BackCallback = s.Opts.Cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Назад",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{s.Opts.B.FirstUserID, s.Opts.B.SecondUserID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, err := s.Opts.Bot.EditMessageText(&telego.EditMessageTextParams{
				MessageID: query.Message.GetMessageID(),
				ChatID:    tu.ID(query.Message.GetChat().ID),
				ParseMode: telego.ModeHTML,
				Text:      s.Label,
				ReplyMarkup: &telego.InlineKeyboardMarkup{
					InlineKeyboard: s.ShopCallbacks,
				},
			})
			if err != nil {
				s.Opts.Log.Sugar().Error(err)
			}
			s.SelectedItem = nil
		},
	})

	s.BuyCallback = s.Opts.Cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Купить",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{s.Opts.B.FirstUserID, s.Opts.B.SecondUserID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			si := s.SelectedItem
			if !s.Opts.B.Score.IsBiggerOrEquals(si.Price) {
				return
			}
			_, err := s.Opts.Bot.EditMessageText(&telego.EditMessageTextParams{
				MessageID: query.Message.GetMessageID(),
				ChatID:    tu.ID(query.Message.GetChat().ID),
				ParseMode: telego.ModeHTML,
				Text:      "Поздравляю, вы купили " + si.Name.String() + " " + strconv.Itoa(si.BuyLevel) + "/" + strconv.Itoa(si.MaxLevel) + " ур.",
			})
			if err != nil {
				s.Opts.Log.Sugar().Error(err)
			}
			s.SelectedItem = nil
		},
	})
}
