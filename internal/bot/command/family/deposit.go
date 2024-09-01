package family

import (
	"famoria/internal/database/mongo/repositories/brak"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/html"
	"fmt"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type depositCmd struct {
	brakRepo brak.Repository
	userRepo user.Repository
	log      *zap.Logger
}

func (c depositCmd) Handle(bot *telego.Bot, update telego.Update) {
	from := update.Message.From
	params := &telego.SendMessageParams{
		ChatID:    tu.ID(update.Message.Chat.ID),
		ParseMode: telego.ModeHTML,
	}
	args := strings.Split(update.Message.Text, " ")

	if len(args) < 2 {
		_, err := bot.SendMessage(params.
			WithText(fmt.Sprintf("%s, укажи сумму для депозита", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	amount, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		// TODO parse exponential (3e3)
		c.log.Sugar().Error(err)
		return
	}

	b, _ := c.brakRepo.FindByUserID(from.ID)
	if b == nil {
		_, err := bot.SendMessage(params.
			WithText(fmt.Sprintf("%s, ты не состоишь в браке. 😥", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	u, _ := c.userRepo.FindByID(from.ID)
	if u == nil {
		return
	}
	if !u.Score.IsBiggerOrEquals(&common.Score{Mantissa: int64(amount)}) {
		_, err := bot.SendMessage(params.
			WithText(fmt.Sprintf("%s, вы ввели сликом большое число для депозита", html.UserMention(from))),
		)
		if err != nil {
			c.log.Sugar().Error(err)
		}
		return
	}

	u.Score.Decrease(amount)
	b.Score.Increase(amount)
	err = c.userRepo.Update(bson.M{"_id": u.OID}, bson.M{"$set": bson.M{"score": u.Score}})
	if err != nil {
		c.log.Sugar().Error(err)
		return
	}
	err = c.brakRepo.Update(bson.M{"_id": b.OID}, bson.M{"$set": bson.M{"score": b.Score}})
	if err != nil {
		c.log.Sugar().Error(err)
		return
	}

	_, err = bot.SendMessage(
		params.WithText(
			fmt.Sprintf("%s, ты успешно внёс депозит в размере %d💰",
				html.UserMention(from), amount),
		),
	)
	if err != nil {
		c.log.Sugar().Error(err)
	}
}
