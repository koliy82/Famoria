package mongo

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"famoria/internal/pkg/proxy"
)

// recordingProxy is a CONNECT proxy that records the target it was asked to
// reach and then answers the client with nothing useful. The handshake will fail;
// what matters is that the driver asked the proxy for the right host.
type recordingProxy struct {
	ln      net.Listener
	mu      sync.Mutex
	targets []string
}

func startRecordingProxy(t *testing.T) *recordingProxy {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	p := &recordingProxy{ln: ln}
	t.Cleanup(func() { _ = ln.Close() })
	go p.serve()
	return p
}

func (p *recordingProxy) serve() {
	for {
		c, err := p.ln.Accept()
		if err != nil {
			return
		}
		go p.handle(c)
	}
}

func (p *recordingProxy) handle(c net.Conn) {
	defer c.Close()

	_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
	req, err := http.ReadRequest(bufio.NewReader(c))
	if err != nil {
		return
	}

	p.mu.Lock()
	p.targets = append(p.targets, req.Method+" "+req.Host)
	p.mu.Unlock()

	// Refuse the tunnel. The point of this test is the routing, not a working
	// MongoDB handshake, and answering 502 makes the driver fail fast instead of
	// leaving the test to time out.
	_, _ = c.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
}

func (p *recordingProxy) seen() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.targets...)
}

func (p *recordingProxy) spec() string {
	return p.ln.Addr().String()
}

// TestDriverRoutesThroughProxyDialer proves the wiring end to end: mongo.Connect
// with our dialer must send a CONNECT for the database host to the proxy rather
// than dialing it directly.
//
// This is the check that matters most, because mongo-driver v1 has no proxy
// option at all — the tunnel is hand-rolled, and a mistake here would leave the
// driver connecting direct and hanging on a blocked network.
func TestDriverRoutesThroughProxyDialer(t *testing.T) {
	p := startRecordingProxy(t)

	ep, err := proxy.Parse(p.spec())
	if err != nil {
		t.Fatal(err)
	}

	const dbHost = "db.example.internal"
	const dbPort = "27017"
	want := dbHost + ":" + dbPort

	client, err := mongo.Connect(
		context.Background(),
		options.Client().
			ApplyURI("mongodb://"+want+"/testdb").
			SetDialer(ep.Dialer()).
			SetConnectTimeout(3*time.Second).
			SetServerSelectionTimeout(3*time.Second),
	)
	if err != nil {
		t.Fatalf("mongo.Connect: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}()

	// Ping forces an actual connection attempt through the dialer.
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = client.Ping(pingCtx, nil) // expected to fail; we only care about routing

	seen := p.seen()
	if len(seen) == 0 {
		t.Fatal("the proxy received no CONNECT, so the driver dialed the database directly")
	}

	found := false
	for _, s := range seen {
		if strings.Contains(s, want) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("proxy saw %v, want a CONNECT for %s", seen, want)
	}
	t.Logf("proxy observed: %v", seen)
}

// TestDialerSignatureMatchesDriverContract pins the interface mongo-driver
// requires, so a future signature change in either library fails here rather
// than at runtime.
func TestDialerSignatureMatchesDriverContract(t *testing.T) {
	ep, err := proxy.Parse("127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}

	var _ options.ContextDialer = ep.Dialer()
}

func TestDeref(t *testing.T) {
	if got := deref(nil); got != "" {
		t.Errorf("deref(nil) = %q, want empty", got)
	}
	s := "user:pass@host:8080"
	if got := deref(&s); got != s {
		t.Errorf("deref(&s) = %q, want %q", got, s)
	}
}
