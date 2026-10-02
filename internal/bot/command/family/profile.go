package family

import (
	"context"
	"fmt"

	"famoria/internal/bot/callback"
	"famoria/internal/bot/callback/static"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/database/mongo/repositories/message"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/common/buttons"
	"famoria/internal/pkg/html"
	"famoria/internal/pkg/i18n"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type profileCmd struct {
	cm           *callback.CallbacksManager
	log          *zap.Logger
	userRepo     user.Repository
	brakRepo     brak.Repository
	messageRepo  message.Repository
	chatSettings chat_settings.Repository
}

func (c profileCmd) Handle(ctx *th.Context, update telego.Update) error {
	from := update.Message.From
	chatID := update.Message.Chat.ID
	lang := c.chatSettings.Lang(chatID)

	// With earnings off, the chat is treated as score-less: no balances are
	// shown and none of the job buttons that award score are rendered.
	earnings := c.chatSettings.IsFeatureEnabled(chatID, chat_settings.FeatEarnings)

	fUser, err := c.userRepo.FindOrUpdate(from)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("🍞🍞 %s 🍞🍞\n", html.Bold(i18n.T(lang, i18n.KeyProfileTitle)))
	text += fmt.Sprintf("👤 %s\n", html.CodeInline(fUser.UsernameOrFull()))
	if earnings {
		text += fmt.Sprintf("💰 %s\n", common.FormattedScore(fUser.Score))
	}

	messageCount, err := c.messageRepo.MessageCount(from.ID, chatID)
	if err == nil {
		text += i18n.T(lang, i18n.KeyProfileMessages, messageCount) + "\n"
	}

	keyboard := buttons.New(3, 3)

	b, err := c.brakRepo.FindByUserID(from.ID, nil)
	if b != nil {
		if b.ChatID == 0 && update.Message.Chat.Type != "private" {
			b.ChatID = chatID
			err = c.brakRepo.Update(bson.M{"_id": b.OID}, bson.M{"$set": bson.M{"chat_id": b.ChatID}})
			if err != nil {
				c.log.Sugar().Error(err)
				return err
			}
		}

		if earnings {
			keyboard.Add(tu.InlineKeyboardButton("🎰").WithCallbackData(static.CasinoData))
			keyboard.Add(tu.InlineKeyboardButton("🐹").WithCallbackData(static.HamsterData))
		}

		tUser, _ := c.userRepo.FindByID(b.PartnerID(fUser.ID))
		text += fmt.Sprintf("\n❤️‍🔥❤️‍🔥      %s      ️‍❤️‍🔥❤️‍🔥\n", html.Bold(i18n.T(lang, i18n.KeyProfileMarriage)))
		if tUser != nil {
			text += fmt.Sprintf("🫂 %s [%s]\n", html.CodeInline(tUser.UsernameOrFull()), b.Duration())
		}

		if b.BabyUserID != nil {
			if earnings {
				keyboard.Add(tu.InlineKeyboardButton("🍼").WithCallbackData(static.GrowKidData))
			}
			bUser, err := c.userRepo.FindByID(*b.BabyUserID)
			if err == nil {
				text += fmt.Sprintf("👼 %s [%s]\n", html.CodeInline(bUser.UsernameOrFull()), b.DurationKid())
			}
		}

		if b.IsSub() {
			days := b.SubDaysCount()
			text += html.Bold(i18n.T(lang, i18n.KeyProfileSubDays, days, i18n.Days(lang, int64(days)))) + "\n"
			if earnings {
				keyboard.Add(tu.InlineKeyboardButton("💎").WithCallbackData(static.AnubisData))
			}
		} else {
			text += i18n.T(lang, i18n.KeyProfileNoSub) + "\n"
		}

		if earnings {
			keyboard.Add(tu.InlineKeyboardButton("💸⛏️👷").WithCallbackData(static.MiningData))
			text += fmt.Sprintf("💰 %v\n", common.FormattedScore(b.Score))
		}
	}

	params := &telego.SendMessageParams{
		ChatID:              tu.ID(chatID),
		ParseMode:           telego.ModeHTML,
		Text:                text,
		DisableNotification: true,
	}

	if len(keyboard.Buttons) != 0 {
		params.ReplyMarkup = keyboard.Build()
	}

	_, err = ctx.Bot().SendMessage(context.Background(), params)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
