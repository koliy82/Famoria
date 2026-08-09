package link

import (
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/database/mongo/repositories/chat_settings"
	"famoria/internal/pkg/html"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

// isChatAdmin checks whether the user is an administrator or creator of the
// chat via the Telegram API. Used to gate the /settings command.
func isChatAdmin(ctx context.Context, bot *telego.Bot, chatID, userID int64) bool {
	member, err := bot.GetChatMember(ctx, &telego.GetChatMemberParams{
		ChatID: tu.ID(chatID),
		UserID: userID,
	})
	if err != nil {
		return false
	}
	status := member.MemberStatus()
	return status == telego.MemberStatusCreator || status == telego.MemberStatusAdministrator
}

// toggleLabel returns the inline-button text for the current state.
func toggleLabel(enabled bool) string {
	if enabled {
		return "🎬 Видео-конвертер: ВКЛ"
	}
	return "🎬 Видео-конвертер: ВЫКЛ"
}

type settingsCmd struct {
	bot          *telego.Bot
	log          *zap.Logger
	chatSettings chat_settings.Repository
	cm           *callback.CallbacksManager
}

func (c settingsCmd) Handle(ctx *th.Context, update telego.Update) error {
	msg := update.Message
	chatID := msg.Chat.ID

	if !isChatAdmin(context.Background(), ctx.Bot(), chatID, msg.From.ID) {
		_, err := ctx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID: tu.ID(chatID),
			Text:   "Команда доступна только администраторам чата.",
			ReplyParameters: &telego.ReplyParameters{
				MessageID:                msg.MessageID,
				AllowSendingWithoutReply: true,
			},
		})
		return err
	}

	enabled := c.chatSettings.IsVideoConverterEnabled(chatID)
	btn := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:      toggleLabel(enabled),
		CtxType:    callback.Temporary,
		OwnerIDs:   []int64{msg.From.ID},
		Time:       time.Minute * 30,
		AnswerText: "Настройка обновлена",
		Callback:   c.makeToggleCallback(chatID, msg.From.ID),
	})

	_, err := ctx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID:    tu.ID(chatID),
		ParseMode: telego.ModeHTML,
		Text: html.Bold("Настройки чата") + "\n\n" +
			"Видео-конвертер: бот проверяет ссылки на видео (YouTube/Shorts/TikTok/Reels и др.), " +
			"скачивает и отправляет видео в чат, а оригинальную ссылку удаляет.",
		ReplyMarkup: tu.InlineKeyboard(tu.InlineKeyboardRow(btn.Inline())),
	})
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}

// makeToggleCallback returns a callback that flips the converter flag and
// refreshes the button label on the originating settings message.
func (c settingsCmd) makeToggleCallback(chatID, userID int64) func(query telego.CallbackQuery) {
	return func(query telego.CallbackQuery) {
		// Re-check admin status at click time to stay safe if privileges changed.
		if !isChatAdmin(context.Background(), c.bot, chatID, userID) {
			_ = c.bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            "Команда доступна только администраторам чата.",
				ShowAlert:       true,
			})
			return
		}

		newState := !c.chatSettings.IsVideoConverterEnabled(chatID)
		c.chatSettings.SetVideoConverter(chatID, newState)

		btn := c.cm.DynamicCallback(callback.DynamicOpts{
			Label:      toggleLabel(newState),
			CtxType:    callback.Temporary,
			OwnerIDs:   []int64{userID},
			Time:       time.Minute * 30,
			AnswerText: "Настройка обновлена",
			Callback:   c.makeToggleCallback(chatID, userID),
		})

		_, err := c.bot.EditMessageReplyMarkup(context.Background(), &telego.EditMessageReplyMarkupParams{
			ChatID:      tu.ID(query.Message.GetChat().ID),
			MessageID:   query.Message.GetMessageID(),
			ReplyMarkup: tu.InlineKeyboard(tu.InlineKeyboardRow(btn.Inline())),
		})
		if err != nil {
			c.log.Sugar().Error(err)
		}
	}
}
