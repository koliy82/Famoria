package database

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go_tg_bot/internal/repositories/user"
)

type Service struct {
	users *user.Ch
}

func Repository(conn driver.Conn) Service {
	var service Service

	service.users = user.New(conn)

	return service
}
