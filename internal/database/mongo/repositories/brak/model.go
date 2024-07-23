package brak

import "github.com/ClickHouse/clickhouse-go"

type Brak struct {
	UUID           clickhouse.UUID      `ch:"uuid"`
	FirstUserID    int64                `ch:"first_user_id"`
	SecondUserID   int64                `ch:"second_user_id"`
	CreateDate     clickhouse.DateTime  `ch:"create_date"`
	BabyUserID     *int64               `ch:"baby_user_id"`
	BabyCreateDate *clickhouse.DateTime `ch:"baby_create_date"`
	Score          int64                `ch:"score"`
}
