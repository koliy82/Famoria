package steam

import (
	"bytes"
	"context"
	"famoria/internal/bot/callback"
	"famoria/internal/bot/handler/waiter"
	"famoria/internal/database/steamapi/repositories/steam_accounts"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/common/buttons"
	"famoria/internal/pkg/html"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

type listCmd struct {
	cm  *callback.CallbacksManager
	mw  *waiter.MessageWaiter
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
		ChatID: update.Message.Chat.ChatID(),
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.MessageID,
			AllowSendingWithoutReply: true,
		},
		DisableNotification: true,
	}
	fromIDs := []int64{update.Message.From.ID}
	keyboard := buttons.New(5, 3)
	AddAccountCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:      "Добавить аккаунт ➕",
		CtxType:    callback.OneClick,
		OwnerIDs:   fromIDs,
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
			Label:    account.Name(),
			CtxType:  callback.ChooseOne,
			OwnerIDs: fromIDs,
			Time:     time.Duration(30) * time.Minute,
			Callback: func(query telego.CallbackQuery) {
				text = "Аккаунт " + html.Bold(account.Name()) + "\n"
				text += "Статус: " + account.PersonaState.String() + "\n"
				text += "Id игр: " + account.Games() + "\n"
				text += "Статус фарма: " + common.Ternary(account.IsFarming, "Запущен", "Остановлен")

				_, err := ctx.Bot().EditMessageText(context.Background(),
					&telego.EditMessageTextParams{
						ChatID:      tu.ID(update.Message.Chat.ID),
						ParseMode:   telego.ModeHTML,
						Text:        text,
						MessageID:   query.Message.GetMessageID(),
						ReplyMarkup: c.accountButtons(account, ctx, fromIDs).InlineBuild(),
					},
				)

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

func (c listCmd) accountButtons(account *steam_accounts.SteamAccount, ctx *th.Context, fromIDs []int64) *buttons.Builder {
	keyboard := buttons.New(1, 5)
	editStatusCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Изменить статус",
		CtxType:  callback.OneClick,
		OwnerIDs: fromIDs,
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			kb2 := buttons.New(2, 3)
			for _, state := range steam_accounts.AvailableStates() {
				if state == account.PersonaState {
					continue
				}
				selectCallback := c.cm.DynamicCallback(callback.DynamicOpts{
					Label:    state.String(),
					CtxType:  callback.ChooseOne,
					OwnerIDs: fromIDs,
					Time:     time.Duration(30) * time.Minute,
					Callback: func(query telego.CallbackQuery) {
						err := c.api.UpdateStatus(account.ID.Hex(), state)
						if err != nil {
							return
						}
						_, err = ctx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
							ChatID:    query.Message.GetChat().ChatID(),
							Text:      "Статус аккаунта теперь: " + state.String(),
							MessageID: query.Message.GetMessageID(),
						})
						if err != nil {
							c.log.Error("failed to edit update status message text", zap.Error(err))
						}
					},
				})
				kb2.Add(selectCallback.Inline())
			}
			_, err := ctx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
				ChatID:      query.Message.GetChat().ChatID(),
				Text:        "Статус аккаунта " + account.Name() + ": " + account.PersonaState.String() + "\nВыберите новый Steam-статус.",
				MessageID:   query.Message.GetMessageID(),
				ReplyMarkup: kb2.InlineBuild(),
			})
			if err != nil {
				c.log.Error("failed to edit update status message text", zap.Error(err))
			}
		},
	})
	gamesCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Изменить игры",
		CtxType:  callback.OneClick,
		OwnerIDs: fromIDs,
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			_, err := ctx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
				ChatID:    query.Message.GetChat().ChatID(),
				MessageID: query.Message.GetMessageID(),
				Text:      "Текущие игры: " + account.Games() + "\nОтправьте игры для фарма через запятую, например: 541, 444, ..",
			})
			if err != nil {
				c.log.Error("failed to edit update game ids text", zap.Error(err))
			}
			// wait for user's next message in the same chat
			c.mw.WaitForMessage(query.From.ID, query.Message.GetChat().ID, 0, time.Duration(30)*time.Minute, func(msgCtx *th.Context, msg telego.Message) {
				// parse text from message (Text or Caption)
				input := ""
				if msg.Text != "" {
					input = msg.Text
				} else if msg.Caption != "" {
					input = msg.Caption
				}
				input = strings.Replace(input, " ", "", -1)
				strIds := strings.Split(input, ",")
				newIds := make([]any, 0, len(strIds))
				for _, s := range strIds {
					if s == "" {
						continue
					}
					id, err := strconv.ParseUint(s, 10, 64)
					if err != nil {
						newIds = append(newIds, s)
						continue
					}
					newIds = append(newIds, id)
				}

				if slices.EqualFunc(account.GameIDs, newIds, func(a any, a2 any) bool { return a == a2 }) {
					_, _ = msgCtx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
						ChatID: msg.Chat.ChatID(),
						Text:   "новые id == старые айди, дурак?",
					})
					return
				}
				err = c.api.UpdateGames(account.ID.Hex(), newIds)
				if err != nil {
					c.log.Error("failed to edit update game ids text", zap.Error(err))
					return
				}
				account.GameIDs = newIds
				_, _ = msgCtx.Bot().SendMessage(context.Background(), &telego.SendMessageParams{
					ChatID: msg.Chat.ChatID(),
					Text:   "новые id: " + account.Games(),
				})
			})
		},
	})
	deleteCallback := c.cm.DynamicCallback(callback.DynamicOpts{
		Label:    "Удалить",
		CtxType:  callback.OneClick,
		OwnerIDs: fromIDs,
		Time:     time.Duration(30) * time.Minute,
		Callback: func(query telego.CallbackQuery) {
			err := c.api.DeleteAccount(account.ID.Hex())
			text := "Аккаунт " + account.Name() + " успешно удалён."
			if err != nil {
				text = "Ошибка при удалении аккаунта"
			}
			_, err = ctx.Bot().EditMessageText(context.Background(), &telego.EditMessageTextParams{
				ChatID:      query.Message.GetChat().ChatID(),
				Text:        text,
				MessageID:   query.Message.GetMessageID(),
				ReplyMarkup: keyboard.InlineBuild(),
			})
			if err != nil {
				c.log.Error("failed to edit delete account message text", zap.Error(err))
			}
		},
	})
	keyboard.Add(editStatusCallback.Inline())
	keyboard.Add(gamesCallback.Inline())
	keyboard.Add(deleteCallback.Inline())
	return keyboard
}
