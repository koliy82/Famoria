package family

import (
	"context"
	"errors"
	"famoria/internal/bot/callback"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/html"
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type withdrawCmd struct {
	cm       *callback.CallbacksManager
	brakRepo brak.Repository
	userRepo user.Repository
	log      *zap.Logger
}

func (c withdrawCmd) Handle(ctx *th.Context, update telego.Update) error {
	from := update.Message.From
	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
	}
	args := strings.Split(update.Message.Text, " ")

	if len(args) < 2 {
		_, err := ctx.Bot().SendMessage(context.Background(), params.
			WithText(fmt.Sprintf("%s, укажи сумму для вывода", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return err
	}

	amount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		c.log.Sugar().Warn(err)
		return err
	}

	b, _ := c.brakRepo.FindByUserID(from.ID, nil)
	if b == nil {
		_, err := ctx.Bot().SendMessage(context.Background(), params.
			WithText(fmt.Sprintf("%s, ты не состоишь в браке. 😥", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return err
	}

	u, _ := c.userRepo.FindByID(from.ID)
	if u == nil {
		return errors.New("user not found")
	}
	if amount <= 0 || b.Score < amount {
		_, err := ctx.Bot().SendMessage(context.Background(), params.
			WithText(fmt.Sprintf("%s, вы ввели сликом большое число для вывода", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return err
	}
	err = c.brakRepo.Update(bson.M{"_id": b.OID}, bson.M{"$inc": bson.M{"score": -amount}})
	if err != nil {
		c.log.Sugar().Warn("Parse withdraw error from user: ", from.ID, " err:", err)
		return err
	}
	err = c.userRepo.Update(bson.M{"_id": u.OID}, bson.M{"$inc": bson.M{"score": amount}})
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}

	_, err = ctx.Bot().SendMessage(
		context.Background(),
		params.WithText(
			fmt.Sprintf("%s, ты успешно вывел из брака %d💰",
				html.UserMention(from), amount),
		),
	)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
