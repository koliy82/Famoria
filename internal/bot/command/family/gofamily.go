package family

import (
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/callback"
)

type goFamily struct {
	cm  *callback.CallbacksManager
	log *zap.Logger
}

func (g goFamily) Handle(bot *telego.Bot, update telego.Update) {
	fUser := update.Message.From
	reply := update.Message.ReplyToMessage

	if reply == nil {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   fmt.Sprintf("@%s, ответь на любое сообщение партнёра. 😘💬", update.Message.From.Username),
			ReplyParameters: &telego.ReplyParameters{
				MessageID: update.Message.GetMessageID(),
			},
		})
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	tUser := reply.From
	if tUser.ID == fUser.ID {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   fmt.Sprintf("@%s, брак с собой нельзя, придётся искать пару. 😥", update.Message.From.Username),
			ReplyParameters: &telego.ReplyParameters{
				MessageID: update.Message.GetMessageID(),
			},
		})
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	if tUser.IsBot {
		_, err := bot.SendMessage(&telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   fmt.Sprintf("@%s, бота не трогай. 👿", update.Message.From.Username),
			ReplyParameters: &telego.ReplyParameters{
				MessageID: update.Message.GetMessageID(),
			},
		})
		if err != nil {
			g.log.Sugar().Error(err)
		}
		return
	}

	//fbrak, err := g.brakRepo.FindByUserID(fUser.ID)
	//
	//if err != nil {
	//	g.log.Sugar().Error(err)
	//	return
	//}
	//
	//if fbrak != nil {
	//	_, err := bot.SendMessage(&telego.SendMessageParams{
	//		ChatID: tu.ID(update.Message.Chat.ID),
	//		Text:   fmt.Sprintf("@%s, у вас уже есть брак! 💍", update.Message.From.Username),
	//		ReplyParameters: &telego.ReplyParameters{
	//			MessageID: update.Message.GetMessageID(),
	//		},
	//	})
	//	if err != nil {
	//		g.log.Sugar().Error(err)
	//	}
	//	return
	//}

	// TODO if fUser not brak

	// TODO if tUser not brak

	yesCallback := g.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Да!❤️‍🔥",
		CtxType:  callback.ChooseOne,
		OwnerIDs: []int64{update.Message.From.ID},
		Time:     5,
		Callback: func(query telego.CallbackQuery) {
			_, err := bot.SendMessage(tu.Messagef(
				telego.ChatID{ID: query.Message.GetChat().ID},
				"Hello %s!", query.From.FirstName,
			))
			if err != nil {
				g.log.Sugar().Error(err)
				return
			}
		},
	})

	noCallback := g.cm.DynamicCallback(callback.DynamicOpts{
		Label:      "Нет!💔",
		CtxType:    callback.ChooseOne,
		OwnerIDs:   []int64{update.Message.From.ID},
		Time:       5,
		AnswerText: "Отказ 🖤",
		Callback: func(query telego.CallbackQuery) {
			_, err := bot.SendMessage(&telego.SendMessageParams{
				ChatID: tu.ID(update.Message.Chat.ID),
				Text:   "Отказ 🖤",
				ReplyParameters: &telego.ReplyParameters{
					MessageID: query.Message.GetMessageID(),
				},
			})
			if err != nil {
				g.log.Sugar().Error(err)
				return
			}
		},
	})

	from := update.Message.From
	_, _ = bot.SendMessage(
		tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"💍 @%s, минуточку внимания.\n"+
				"💖 @%s сделал вам предложение руки и сердца.",
			from.Username,
			from.Username,
		).WithReplyMarkup(
			tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					yesCallback.Inline(),
					noCallback.Inline(),
				),
			),
		))
}
