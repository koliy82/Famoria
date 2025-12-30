package shop

import (
	"context"
	"errors"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/inventory"
	"famoria/internal/database/mongo/repositories/brak"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
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
	B        *brak.Brak
	Params   *telego.SendMessageParams
	BotCtx   *th.Context
	Manager  *item.Manager
	Log      *zap.Logger
	Cm       *callback.CallbacksManager
	BrakRepo brak.Repository
}

func New(opts *Opts) (*Shop, error) {
	s := &Shop{
		CurrentPage: 1,
		MaxRows:     3,
		MaxCells:    4,
		Opts:        opts,
	}
	err := s.UpdateShop()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Shop) CurrentButtonsPage() [][]telego.InlineKeyboardButton {
	s.Label = fmt.Sprintf("Потайная лавка (%d/%d стр.)\n", s.CurrentPage, s.MaxPages)
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
				Label:    si.Emoji,
				CtxType:  callback.Temporary,
				OwnerIDs: []int64{s.Opts.B.FirstUserID, s.Opts.B.SecondUserID},
				Time:     time.Duration(30) * time.Minute,
				Callback: func(query telego.CallbackQuery) {
					s.SelectedItem = si
					_, err := s.Opts.BotCtx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
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
					_ = s.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(),
						&telego.AnswerCallbackQueryParams{
							CallbackQueryID: query.ID,
						})
				},
			})
			s.ShopCallbacks[i] = append(s.ShopCallbacks[i], dCallback.Inline())
		}
	}
	s.Label += fmt.Sprintf("Доступные средства - %s 💰", s.Opts.B.Score.GetFormattedScore())

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
			_, err := s.Opts.BotCtx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
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
			_ = s.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
			})
		}})

	s.NavigateNext = s.Opts.Cm.DynamicCallback(callback.DynamicOpts{
		Label:    "➡️",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{s.Opts.B.FirstUserID, s.Opts.B.SecondUserID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, err := s.Opts.BotCtx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
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
			_ = s.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
			})
		}})
}

func (s *Shop) UpdateShop() error {
	s.Items = s.Opts.B.GetAvailableItems(s.Opts.Manager)
	sort.Slice(s.Items, func(i, j int) bool {
		if s.Items[i].Price.Exponent == s.Items[j].Price.Exponent {
			return s.Items[i].Price.Mantissa < s.Items[j].Price.Mantissa
		}
		return s.Items[i].Price.Exponent < s.Items[j].Price.Exponent
	})
	if len(s.Items) == 0 {
		_, err := s.Opts.BotCtx.Bot().SendMessage(context.Background(), s.Opts.Params.
			WithText("Вы скупили все доступные предметы на данный момент, милорд."),
		)
		if err != nil {
			s.Opts.Log.Sugar().Error(err)
		}
		return errors.New("no items in shop")
	}
	s.MaxPages = len(s.Items) / (s.MaxRows * s.MaxCells)
	if len(s.Items)%(s.MaxRows*s.MaxCells) != 0 {
		s.MaxPages++
	}
	s.SetItemCallbacks()
	if s.MaxPages > 1 {
		s.SetNavigateButtons()
	}
	s.CurrentButtonsPage()
	return nil
}

func (s *Shop) SetItemCallbacks() {
	s.BackCallback = s.Opts.Cm.DynamicCallback(callback.DynamicOpts{
		Label:    "⬅️ Назад",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{s.Opts.B.FirstUserID, s.Opts.B.SecondUserID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, err := s.Opts.BotCtx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
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
			_ = s.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
			})
		},
	})

	s.BuyCallback = s.Opts.Cm.DynamicCallback(callback.DynamicOpts{
		Label:    "💳 Купить",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{s.Opts.B.FirstUserID, s.Opts.B.SecondUserID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			answerParams := &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
			}
			si := s.SelectedItem
			actualBrak, err := s.Opts.BrakRepo.FindByUserID(s.Opts.B.FirstUserID, s.Opts.Manager)
			if actualBrak.Events.Shop.Sale > 0 {
				si.Price = si.Price.GetSaleScore(actualBrak.Events.Shop.Sale)
			}
			if !actualBrak.Score.IsBiggerOrEquals(si.Price) {
				_ = s.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), answerParams.
					WithText("У вас недостаточно средств для покупки/улучшения предмета."),
				)
				return
			}
			ii, ok := actualBrak.Inventory.Items[si.Name]
			if !ok {
				ii = inventory.Item{
					Id:           si.Name,
					CurrentLevel: 0,
				}
			}
			buyItem := ii.GetItem(s.Opts.Manager)
			if si.BuyLevel >= si.MaxLevel || ii.CurrentLevel >= buyItem.MaxLevel {
				_ = s.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), answerParams.
					WithText("Предмет уже максимального уровня."),
				)
				return
			}
			actualBrak.Score.Minus(si.Price)
			ii.CurrentLevel++
			actualBrak.Inventory.Items[si.Name] = ii
			err = s.Opts.BrakRepo.Update(bson.M{"_id": actualBrak.OID},
				bson.M{"$set": bson.M{
					"score":     actualBrak.Score,
					"inventory": actualBrak.Inventory},
				})
			if err != nil {
				s.Opts.Log.Sugar().Error(err)
				_ = s.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), answerParams.
					WithText("Произошла ошибка при покупке/улучшении предмета."),
				)
			}
			// Update item in shop list
			s.Opts.B = actualBrak
			for i, ui := range s.Items {
				if ui.Name == si.Name {
					s.Items[i].BuyLevel++
				}
			}
			err = s.UpdateShop()
			_, err = s.Opts.BotCtx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
				MessageID: query.Message.GetMessageID(),
				ChatID:    tu.ID(query.Message.GetChat().ID),
				ParseMode: telego.ModeHTML,
				Text:      "Поздравляю, вы купили " + si.Name.String() + " " + strconv.Itoa(si.BuyLevel) + "/" + strconv.Itoa(si.MaxLevel) + " ур.",
				ReplyMarkup: tu.InlineKeyboard(
					tu.InlineKeyboardRow(s.BackCallback.Inline()),
				),
			})
			if err != nil {
				s.Opts.Log.Sugar().Error(err)
			}
			s.SelectedItem = nil
			_ = s.Opts.BotCtx.Bot().AnswerCallbackQuery(context.Background(), answerParams)
		},
	})
}
