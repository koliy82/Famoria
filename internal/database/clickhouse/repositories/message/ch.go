package message

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"
)

type Ch struct {
	conn driver.Conn
	log  *zap.Logger
}

func (c Ch) Insert(message *Message) {
	sql, args, err := sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Insert("koliy82.message").
		Columns("id", "chat_id", "user_id", "date", "text", "caption", "reply_id").
		Values(message.ID, message.ChatID, message.UserID, message.Date, message.Text, message.Caption, message.ReplyID).
		ToSql()
	err = c.conn.AsyncInsert(context.Background(), sql, false, args...)
	if err != nil {
		c.log.Error("Error inserting message", zap.Error(err))
	}
	c.log.Sugar().Debug("New Message: ", message)
}

func New(conn driver.Conn, log *zap.Logger) *Ch {
	return &Ch{
		conn: conn,
		log:  log,
	}
}
