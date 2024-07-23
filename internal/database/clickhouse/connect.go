package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go_tg_bot/internal/config"
	"net"
	"time"
)

func New(cfg config.Config) driver.Conn {
	var (
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
			Debug: true,
			Debugf: func(format string, v ...interface{}) {
				fmt.Printf(format, v)
			},
			DialTimeout:      time.Duration(10) * time.Second,
			MaxOpenConns:     5,
			MaxIdleConns:     5,
			ConnMaxLifetime:  time.Duration(10) * time.Minute,
			ConnOpenStrategy: clickhouse.ConnOpenInOrder,
			BlockBufferSize:  10,
			DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "tcp", addr)
			},
		})
	)

	if err != nil {
		panic(err)
	}

	defer func(conn driver.Conn) {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}(conn)

	if err := conn.Ping(context.Background()); err != nil {
		var exception *clickhouse.Exception
		if errors.As(err, &exception) {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
			panic(err)
		}
	}
	return conn
}
