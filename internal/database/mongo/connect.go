package mongo

import (
	"context"
	"fmt"
	"time"

	"famoria/internal/config"
	"famoria/internal/pkg/proxy"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// connectTimeout bounds the initial connection handshake. Without it, an
// unreachable proxy or host makes startup hang indefinitely instead of failing
// with a usable error.
const connectTimeout = 30 * time.Second

func New(cfg config.Config) *mongo.Client {
	opts := options.Client().
		ApplyURI(cfg.MongoURI).
		SetConnectTimeout(connectTimeout).
		SetServerSelectionTimeout(connectTimeout)

	// MongoDB speaks its own binary wire protocol, not HTTP, so it cannot be
	// routed through an http.Transport. It goes through a raw CONNECT tunnel
	// instead; the driver still performs its own TLS inside that tunnel when the
	// URI asks for it.
	if cfg.DBProxy {
		ep := proxy.MustParse(deref(cfg.DBProxyURL))
		opts = opts.SetDialer(ep.Dialer())
		// The zap logger is not available here: mongo.New runs before SetupLogger,
		// so this goes to stdout, which docker logs captures.
		fmt.Printf("mongo: database traffic routed through proxy %s\n", ep.Masked())
	}

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	return client
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
