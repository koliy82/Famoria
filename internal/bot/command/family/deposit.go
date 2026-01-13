package family

import (
	"context"
	"errors"
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/common"
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

type depositCmd struct {
	brakRepo brak.Repository
	userRepo user.Repository
	log      *zap.Logger
}

func (c depositCmd) Handle(ctx *th.Context, update telego.Update) error {
	from := update.Message.From
	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
	}
	args := strings.Split(update.Message.Text, " ")

	if len(args) < 2 {
		_, err := ctx.Bot().SendMessage(context.Background(), params.
			WithText(fmt.Sprintf("%s, укажи сумму для депозита", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return err
	}

	amount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		// TODO parse exponential (3e3)
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
	if !u.Score.IsBiggerOrEquals(&common.Score{Mantissa: int64(amount)}) {
		_, err := ctx.Bot().SendMessage(context.Background(), params.
			WithText(fmt.Sprintf("%s, вы ввели сликом большое число для депозита", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return err
	}

	u.Score.Decrease(int64(amount))
	b.Score.Increase(int64(amount))
	err = c.userRepo.Update(bson.M{"_id": u.OID}, bson.M{"$set": bson.M{"score": u.Score}})
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}
	err = c.brakRepo.Update(bson.M{"_id": b.OID}, bson.M{"$set": bson.M{"score": b.Score}})
	if err != nil {
		c.log.Sugar().Error(err)
		return err
	}

	_, err = ctx.Bot().SendMessage(context.Background(),
		params.WithText(
			fmt.Sprintf("%s, ты успешно внёс депозит в размере %d💰",
				html.UserMention(from), amount),
		),
	)
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return err
}
