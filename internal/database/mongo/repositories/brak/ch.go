package brak

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
)

type Ch struct {
	conn driver.Conn
}

func (c *Ch) FindByUserID(id int64) (*Brak, error) {

	sql, args, err := sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("*").
		From("koliy82.family").
		Where(sq.Or{
			sq.Eq{"first_user_id": id},
			sq.Eq{"second_user_id": id},
		}).
		Limit(1).
		ToSql()

	brak := &Brak{}

	err = c.conn.QueryRow(context.Background(), sql, args...).Scan(
		&brak.UUID,
		&brak.FirstUserID,
		&brak.SecondUserID,
		&brak.CreateDate,
		&brak.BabyUserID,
		&brak.BabyCreateDate,
		&brak.Score,
	)
	if err != nil {
		return nil, err
	}

	return brak, nil
}

func New(conn driver.Conn) *Ch {
	return &Ch{conn: conn}
}
