package steam

import (
	"bytes"
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/database/steamapi/repositories/steam_accounts"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/common/buttons"
	"famoria/internal/pkg/html"
	"fmt"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type listCmd struct {
	cm  *callback.CallbacksManager
	api *steam_accounts.SteamAPI
	log *zap.Logger
}

func (c listCmd) Handle(ctx *th.Context, update telego.Update) error {
	accounts, err := c.api.FindByUserID(update.Message.From.ID)
	if err != nil {
		c.log.Error("failed to fetch steam accounts", zap.Int64("user_id", update.Message.From.ID), zap.Error(err))
		return err
	}
	params := &telego.SendMessageParams{
		ChatID: tu.ID(update.Message.Chat.ID),
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.GetMessageID(),
			AllowSendingWithoutReply: true,
		},
		DisableNotification: true,
	}
	keyboard := buttons.New(5, 3)
	AddAccountCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:      "Добавить аккаунт ➕",
		CtxType:    callback.OneClick,
		OwnerIDs:   []int64{update.Message.From.ID},
		Time:       time.Duration(30) * time.Minute,
		AnswerText: common.Ternary(update.Message.Chat.Type == "private", "", "Сообщение с авторизацией было отправлено в ЛС."),
		Callback: func(query telego.CallbackQuery) {
			qrBytes, err := c.api.GenerateQRCode(update.Message.From.ID)
			if err != nil {
				return
			}
			_, err = ctx.Bot().SendPhoto(context.Background(), &telego.SendPhotoParams{
				ChatID:  tu.ID(update.Message.From.ID),
				Caption: fmt.Sprintf("Отсканируйте QR-код в приложении Steam, либо вручную введите команду /addsteam login123:password228:2FA-guard-code"),
				Photo: tu.File(
					tu.NameReader(
						bytes.NewReader(qrBytes),
						"steam-qr.png",
					),
				),
			})
			if err != nil {
				c.log.Error("failed to send auth message", zap.Error(err))
			}
			// wait user message and validate login:password:guard-code(code optional)
		},
	})

	if len(accounts) == 0 {
		keyboard.Add(AddAccountCallback.Inline())
		_, err = ctx.Bot().SendMessage(ctx,
			params.WithText("У вас ещё не добавлено ни одного Steam аккаунта.").
				WithReplyMarkup(keyboard.Build()),
		)
		return err
	}

	text := "Ваши добавленные Steam аккаунты:\n"
	for i, account := range accounts {
		text += fmt.Sprintf("%d. %s\n", i+1, account.Name())
		userCallback := c.cm.DynamicCallback(callback.DynamicOpts{
			Label:    account.ID.String(),
			CtxType:  callback.ChooseOne,
			OwnerIDs: []int64{update.Message.From.ID},
			Time:     time.Duration(30) * time.Minute,
			Callback: func(query telego.CallbackQuery) {
				text = "Аккаунт " + html.Bold(account.Name()) + "\n"
				text += "Статус: " + account.PersonaState.String() + "\n"
				text += "Id игр: " + strings.Trim(strings.Replace(fmt.Sprint(account.GameIDs), " ", ", ", -1), "[]") + "\n"
				text += "Статус фарма: " + common.Ternary(account.IsFarming, "Запущен", "Остановлен")
				_, err := ctx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
					ChatID:    tu.ID(update.Message.Chat.ID),
					MessageID: query.Message.GetMessageID(),
					Text:      text,
					ParseMode: telego.ModeHTML,
				})
				if err != nil {
					c.log.Error("failed to edit message text", zap.Error(err))
				}
			},
		})
		keyboard.Add(userCallback.Inline())
	}
	keyboard.Add(AddAccountCallback.Inline())
	_, err = ctx.Bot().SendMessage(ctx,
		params.WithText(text).
			WithReplyMarkup(keyboard.Build()),
	)
	return err
}
