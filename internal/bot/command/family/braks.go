package family

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"famoria/internal/bot/callback"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/html"
	"famoria/internal/pkg/i18n"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type pagesCmd struct {
	cm           *callback.CallbacksManager
	brakRepo     brak.Repository
	chatSettings chat_settings.Repository
	isLocal      bool
	log          *zap.Logger
}

func (c pagesCmd) Handle(ctx *th.Context, update telego.Update) error {
	var page int64 = 1
	var limit int64 = 5
	var keyboard *telego.InlineKeyboardMarkup
	var header string
	var filter bson.M
	var pages int64

	chatID := update.Message.Chat.ID
	lang := c.chatSettings.Lang(chatID)

	// In chats with earnings turned off, a score ranking is meaningless (no score
	// is accumulated or shown), so the list falls back to creation date.
	earnings := c.chatSettings.IsFeatureEnabled(chatID, chat_settings.FeatEarnings)
	sort := brak.BrakSortByScore
	if !earnings {
		sort = brak.BrakSortByCreateDate
	}

	params := &telego.SendMessageParams{
		ChatID:    tu.ID(chatID),
		ParseMode: telego.ModeHTML,
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.GetMessageID(),
			AllowSendingWithoutReply: true,
		},
		DisableNotification: true,
	}

	if c.isLocal {
		filter = bson.M{"chat_id": chatID}
	} else {
		filter = bson.M{}
	}

	braks, count, err := c.brakRepo.FindBraksByPage(page, limit, filter, sort)
	if err != nil {
		_, err = ctx.Bot().SendMessage(context.Background(), params.WithText(i18n.T(lang, i18n.KeyBraksError)))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return err
	}

	pages = int64(math.Ceil(float64(count) / float64(limit)))

	header = i18n.Marriages(lang, count, !c.isLocal)

	// renderPage rebuilds the page body for the current data. It is shared by the
	// first render and by both paging buttons.
	renderPage := func() string {
		return header + fillPage(braks, page, limit, lang, earnings)
	}

	backCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "⬅️",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{update.Message.From.ID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			if page == 1 {
				page = pages
			} else {
				page--
			}

			braks, count, err = c.brakRepo.FindBraksByPage(page, limit, filter, sort)
			if err != nil {
				return
			}
			header = i18n.Marriages(lang, count, !c.isLocal)

			keyboard.InlineKeyboard[0][1].Text = strconv.FormatInt(page, 10)
			_, err = ctx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
				MessageID:   query.Message.GetMessageID(),
				ChatID:      tu.ID(chatID),
				ParseMode:   telego.ModeHTML,
				Text:        renderPage(),
				ReplyMarkup: keyboard,
			})
			if err != nil {
				c.log.Sugar().Error(err)
			}
		},
	})

	currentCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    strconv.FormatInt(page, 10),
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{update.Message.From.ID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_ = ctx.Bot().AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            i18n.T(lang, i18n.KeyBraksPageHint, page),
			})
		},
	})

	nextCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "➡️",
		CtxType:  callback.Temporary,
		OwnerIDs: []int64{update.Message.From.ID},
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			if page == pages {
				page = 1
			} else {
				page++
			}

			braks, count, err = c.brakRepo.FindBraksByPage(page, limit, filter, sort)
			if err != nil {
				return
			}
			header = i18n.Marriages(lang, count, !c.isLocal)

			keyboard.InlineKeyboard[0][1].Text = strconv.FormatInt(page, 10)
			_, err = ctx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
				MessageID:   query.Message.GetMessageID(),
				ChatID:      tu.ID(chatID),
				ParseMode:   telego.ModeHTML,
				Text:        renderPage(),
				ReplyMarkup: keyboard,
			})
			if err != nil {
				c.log.Sugar().Error(err)
			}
		},
	})

	keyboard = tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			backCallback.Inline(),
			currentCallback.Inline(),
			nextCallback.Inline(),
		),
	)

	_, err = ctx.Bot().SendMessage(context.Background(), params.
		WithText(renderPage()).
		WithReplyMarkup(keyboard).
		WithDisableNotification(),
	)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}

// fillPage renders one page of the marriage list. Balances are omitted when the
// chat has earnings disabled, matching the profile behaviour.
func fillPage(braks []*brak.UsersBrak, page int64, limit int64, lang i18n.Lang, earnings bool) string {
	var text string
	if len(braks) == 0 {
		return i18n.T(lang, i18n.KeyBraksEmpty)
	}
	for index, m := range braks {
		text += fmt.Sprintf("%d. ", index+1+(int(page)-1)*int(limit))

		if m.First == nil {
			text += "?"
		} else {
			text += m.First.UsernameOrFull()
		}

		if m.Brak.IsSub() {
			text += " ❤️‍🔥 "
		} else {
			text += i18n.T(lang, i18n.KeyBraksAnd)
		}

		if m.Second == nil {
			text += "?"
		} else {
			text += m.Second.UsernameOrFull()
		}

		if m.Brak.BabyUserID != nil && m.Baby != nil {
			text += fmt.Sprintf(" 👼 %s",
				html.CodeInline(m.Baby.UsernameOrFull()),
			)
		}

		text += "\n   " + i18n.T(lang, i18n.KeyBraksDuration, m.Brak.Duration())
		if earnings {
			text += fmt.Sprintf(" - %s 💰\n", common.FormattedScore(m.Brak.Score))
		} else {
			text += "\n"
		}
	}
	return text
}
