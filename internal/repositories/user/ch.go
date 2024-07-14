package user

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
)

type Ch struct {
	conn driver.Conn
}

func (c *Ch) FindByID(ctx context.Context, id int64) (*User, error) {

	sql, args, err := sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("*").
		From("koliy82.user").
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()

	user := &User{}

	err = c.conn.QueryRow(ctx, sql, args...).ScanStruct(&user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func New(conn driver.Conn) *Ch {
	return &Ch{conn: conn}
}
