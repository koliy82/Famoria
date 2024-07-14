package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"
	"go_tg_bot/internal/config"
	"os"
)

func New(log *zap.Logger, cfg config.Config) driver.Conn {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickhouseURL, cfg.ClickhousePort)},
			Auth: clickhouse.Auth{
				Database: cfg.ClickhouseDatabase,
				Username: cfg.ClickhouseUser,
				Password: cfg.ClickhousePassword,
			},
			ClientInfo: clickhouse.ClientInfo{
				Products: []struct {
					Name    string
					Version string
				}{
					{Name: "go_tg_bot", Version: "0.1"},
				},
			},
			Debugf: func(format string, v ...interface{}) {
				fmt.Printf(format, v)
			},
		})
	)

	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if err := conn.Ping(ctx); err != nil {
		var exception *clickhouse.Exception
		if errors.As(err, &exception) {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
			log.Error(err.Error())
			os.Exit(1)
		}
	}
	return conn
}
