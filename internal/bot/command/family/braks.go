package family

import (
	"famoria/internal/bot/callback"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/pkg/html"
	"famoria/internal/pkg/plural"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"math"
	"strconv"
	"time"
)

type pagesCmd struct {
	cm       *callback.CallbacksManager
	brakRepo brak.Repository
	isLocal  bool
	log      *zap.Logger
}

func (c pagesCmd) Handle(bot *telego.Bot, update telego.Update) {
	var page int64 = 1
	var limit int64 = 5
	var keyboard *telego.InlineKeyboardMarkup
	var header string
	var filter bson.M
	var pages int64

	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.GetMessageID(),
			AllowSendingWithoutReply: true,
		},
		DisableNotification: true,
	}

	if c.isLocal {
		filter = bson.M{"chat_id": update.Message.Chat.ID}
	} else {
		filter = bson.M{}
	}

	braks, count, err := c.brakRepo.FindBraksByPage(page, limit, filter)

	pages = int64(math.Ceil(float64(count) / float64(limit)))

	if err != nil {
		_, err = bot.SendMessage(params.WithText("Произошла ошибка при получении списка браков"))
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	if c.isLocal {
		header = fmt.Sprintf("💍 %d %s В ГРУППЕ 💍\n",
			count, plural.Declension(count, "БРАК", "БРАКА", "БРАКОВ"),
		)
	} else {
		header = fmt.Sprintf("💍 %d %s В ЧАТАХ 💍\n",
			count, plural.Declension(count, "БРАК", "БРАКА", "БРАКОВ"),
		)
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

			braks, count, err = c.brakRepo.FindBraksByPage(page, limit, filter)
			if err != nil {
				return
			}

			keyboard.InlineKeyboard[0][1].Text = strconv.FormatInt(page, 10)
			_, err = bot.EditMessageText(&telego.EditMessageTextParams{
				MessageID:   query.Message.GetMessageID(),
				ChatID:      tu.ID(update.Message.Chat.ID),
				ParseMode:   telego.ModeHTML,
				Text:        header + fillPage(braks, page, limit),
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
			_ = bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            fmt.Sprintf("Страница №%d (На ней же не написано? =/)", page),
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

			braks, count, err = c.brakRepo.FindBraksByPage(page, limit, filter)
			if err != nil {
				return
			}

			keyboard.InlineKeyboard[0][1].Text = strconv.FormatInt(page, 10)
			_, err = bot.EditMessageText(&telego.EditMessageTextParams{
				MessageID:   query.Message.GetMessageID(),
				ChatID:      tu.ID(update.Message.Chat.ID),
				ParseMode:   telego.ModeHTML,
				Text:        header + fillPage(braks, page, limit),
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

	_, err = bot.SendMessage(params.
		WithText(header + fillPage(braks, page, limit)).
		WithReplyMarkup(keyboard).
		WithDisableNotification(),
	)
	if err != nil {
		c.log.Sugar().Error(err)
	}
}

func fillPage(braks []*brak.UsersBrak, page int64, limit int64) string {
	var text string
	if len(braks) == 0 {
		return "В этом чате нет браков"
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
			text += " и "
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

		text += fmt.Sprintf("\n   ⏳ %s", m.Brak.Duration())
		text += fmt.Sprintf(" - %s 💰\n", m.Brak.Score.GetFormattedScore())
	}
	return text
}
