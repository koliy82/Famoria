package proxy

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestParseAcceptsDocumentedForms(t *testing.T) {
	cases := []struct {
		in         string
		wantScheme string
		wantHost   string
		wantPort   string
		wantUser   string
	}{
		// The form documented for this project: bare credentials, no scheme.
		{"login:pass@1.2.3.4:8080", "http", "1.2.3.4", "8080", "login"},
		{"user:pass@proxy.example.com:3128", "http", "proxy.example.com", "3128", "user"},
		{"http://user:pass@1.2.3.4:8080", "http", "1.2.3.4", "8080", "user"},
		{"https://user:pass@1.2.3.4:8443", "https", "1.2.3.4", "8443", "user"},
		{"  https://user:pass@1.2.3.4:8443  ", "https", "1.2.3.4", "8443", "user"},
		// Passwords are percent-encoded because the value is parsed as a URL.
		{"user:p%40ss%3Aword@1.2.3.4:8080", "http", "1.2.3.4", "8080", "user"},
		// No credentials at all is valid for open proxies.
		{"1.2.3.4:8080", "http", "1.2.3.4", "8080", ""},
	}

	for _, c := range cases {
		ep, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", c.in, err)
			continue
		}
		u := ep.URL()
		if u.Scheme != c.wantScheme {
			t.Errorf("Parse(%q) scheme = %q, want %q", c.in, u.Scheme, c.wantScheme)
		}
		if u.Hostname() != c.wantHost {
			t.Errorf("Parse(%q) host = %q, want %q", c.in, u.Hostname(), c.wantHost)
		}
		if u.Port() != c.wantPort {
			t.Errorf("Parse(%q) port = %q, want %q", c.in, u.Port(), c.wantPort)
		}
		gotUser := ""
		if u.User != nil {
			gotUser = u.User.Username()
		}
		if gotUser != c.wantUser {
			t.Errorf("Parse(%q) user = %q, want %q", c.in, gotUser, c.wantUser)
		}
	}
}

// TestParseDecodesSpecialPassword guards credential handling: a password with
// reserved characters must survive percent-decoding, or authentication fails
// with a confusing 407 rather than a parse error.
func TestParseDecodesSpecialPassword(t *testing.T) {
	ep, err := Parse("https://user:p%40ss%3Aword@1.2.3.4:8443")
	if err != nil {
		t.Fatal(err)
	}
	pass, ok := ep.URL().User.Password()
	if !ok || pass != "p@ss:word" {
		t.Errorf("decoded password = %q (ok=%v), want %q", pass, ok, "p@ss:word")
	}
}

func TestParseRejectsInvalid(t *testing.T) {
	cases := map[string]string{
		"empty":         "",
		"spaces only":   "   ",
		"no port":       "http://user:pass@1.2.3.4",
		"no host":       "http://:8080",
		"bare no port":  "user:pass@host",
		"socks scheme":  "socks5://user:pass@1.2.3.4:1080",
		"file scheme":   "file:///etc/passwd",
		"unparsable":    "http://%zz:8080",
		"scheme only":   "http://",
		"just a scheme": "https://",
		"ipv6 no port":  "http://[::1]",
		"garbage":       "not a url at all",
		"empty scheme":  "://host:8080",
		"ftp scheme":    "ftp://host:21",
		"double scheme": "http://https://host:8080",
		"control chars": "http://host\x00:8080",
	}
	for name, in := range cases {
		if _, err := Parse(in); err == nil {
			t.Errorf("%s: Parse(%q) should fail", name, in)
		}
	}
}

func TestMustParsePanicsOnInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse should panic on an unusable proxy")
		}
	}()
	MustParse("not a url")
}

// TestMaskURLNeverLeaksPassword is the security guard: the proxy URL contains
// credentials and is logged at startup.
func TestMaskURLNeverLeaksPassword(t *testing.T) {
	const secret = "s3cr3t-value"
	inputs := []string{
		"user:" + secret + "@1.2.3.4:8080",
		"https://user:" + secret + "@1.2.3.4:8443",
		"http://user:" + secret + "@host:3128",
	}
	for _, in := range inputs {
		masked := MaskURL(in)
		if strings.Contains(masked, secret) {
			t.Errorf("MaskURL(%q) leaked the password: %q", in, masked)
		}
		if !strings.Contains(masked, "user") {
			t.Errorf("MaskURL(%q) dropped the username: %q", in, masked)
		}

		ep, err := Parse(in)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(ep.Masked(), secret) {
			t.Errorf("Endpoint.Masked() leaked the password: %q", ep.Masked())
		}
	}
}

// TestMaskURLLeavesUnparsableVisible ensures a malformed value is not replaced by
// something unrelated, which would hide the real misconfiguration.
func TestMaskURLLeavesUnparsableVisible(t *testing.T) {
	if got := MaskURL("total garbage"); got != "total garbage" {
		t.Errorf("MaskURL should return unparseable input unchanged, got %q", got)
	}
	if got := MaskURL(""); got != "" {
		t.Errorf("MaskURL(\"\") = %q, want empty", got)
	}
}

func TestEndpointHasAuth(t *testing.T) {
	withAuth, _ := Parse("user:pass@1.2.3.4:8080")
	if !withAuth.hasAuth() {
		t.Error("expected hasAuth for a URL with credentials")
	}

	without, _ := Parse("1.2.3.4:8080")
	if without.hasAuth() {
		t.Error("expected no auth for a URL without credentials")
	}
}

func TestURLReturnsCopy(t *testing.T) {
	ep, err := Parse("user:pass@1.2.3.4:8080")
	if err != nil {
		t.Fatal(err)
	}

	u := ep.URL()
	if u.User == nil {
		t.Fatal("parsed endpoint lost its credentials")
	}
	u.User = nil // mutate the caller's copy

	// The endpoint must be unaffected: URL() hands out a copy precisely so a
	// caller cannot strip the credentials used for proxy authentication.
	fresh := ep.URL()
	if fresh.User == nil {
		t.Error("URL() did not return a copy; mutating it corrupted the endpoint")
	}
	if got, ok := fresh.User.Password(); !ok || got != "pass" {
		t.Errorf("endpoint password = %q (present=%v), want %q", got, ok, "pass")
	}
}

func TestHTTPTransportUsesProxy(t *testing.T) {
	ep, _ := Parse("https://user:pass@proxy.test:9999")
	tr := ep.HTTPTransport()

	if tr.Proxy == nil {
		t.Fatal("transport has no Proxy func")
	}
	u, err := tr.Proxy(&http.Request{URL: &url.URL{Scheme: "https", Host: "api.telegram.org"}})
	if err != nil {
		t.Fatal(err)
	}
	if u == nil || u.Host != "proxy.test:9999" {
		t.Errorf("proxy target = %v, want proxy.test:9999", u)
	}

	// Pooling must stay enabled, and the per-host idle limit raised above Go's
	// default. Note MaxIdleConnsPerHost==0 means "fall back to
	// DefaultMaxIdleConnsPerHost", which is 2 — far too few here, since the bot
	// uses one host from long polling, sends, uploads and the log sink at once,
	// and every evicted connection costs an extra CONNECT round-trip.
	if tr.MaxIdleConns <= 0 {
		t.Errorf("transport lost connection pooling: MaxIdleConns=%d", tr.MaxIdleConns)
	}
	if tr.MaxIdleConnsPerHost <= 2 {
		t.Errorf("MaxIdleConnsPerHost = %d, want above the effective default of 2",
			tr.MaxIdleConnsPerHost)
	}
	if tr.DisableKeepAlives {
		t.Error("keep-alive must stay enabled")
	}
	// A response-header timeout would abort Telegram long polling, which holds a
	// request open for tens of seconds before answering.
	if tr.ResponseHeaderTimeout != 0 {
		t.Errorf("ResponseHeaderTimeout = %v would break long polling", tr.ResponseHeaderTimeout)
	}
}

func TestHTTPClientIsUsable(t *testing.T) {
	ep, _ := Parse("user:pass@proxy.test:9999")
	c := ep.HTTPClient()
	if c == nil || c.Transport == nil {
		t.Fatal("HTTPClient returned an unusable client")
	}
}

// --- live CONNECT tunnel tests -------------------------------------------
//
// These run a real TCP proxy in-process and drive a non-HTTP protocol through
// the tunnel, which is exactly what the MongoDB wire protocol needs.

// fakeProxy is a minimal HTTP CONNECT proxy used to exercise the tunnel.
type fakeProxy struct {
	t          *testing.T
	ln         net.Listener
	requirePwd string // empty disables authentication

	mu        sync.Mutex
	requests  []string
	gotAuth   []string
	connected []string
	rejected  []string
}

// newFakeProxy starts a plaintext CONNECT proxy on a loopback port.
func newFakeProxy(t *testing.T, requireAuth bool) *fakeProxy {
	t.Helper()
	return startFakeProxy(t, requireAuth, nil)
}

// newTLSFakeProxy starts an https:// proxy and returns the certificate it
// serves, so a test can build a matching trust pool. Generating a second
// certificate on the client side would not validate against this one.
func newTLSFakeProxy(t *testing.T, requireAuth bool) (*fakeProxy, tls.Certificate) {
	t.Helper()
	cert, err := selfSignedCert()
	if err != nil {
		t.Fatal(err)
	}
	p := startFakeProxy(t, requireAuth, &tls.Config{Certificates: []tls.Certificate{cert}})
	return p, cert
}

// startFakeProxy builds the listener and starts exactly one accept loop.
//
// The TLS wrapping happens before the loop starts rather than by swapping the
// listener afterwards: two loops accepting on the same underlying listener would
// race, and connections would nondeterministically be handled as plaintext or
// TLS.
func startFakeProxy(t *testing.T, requireAuth bool, tlsCfg *tls.Config) *fakeProxy {
	t.Helper()

	p := &fakeProxy{t: t}
	if requireAuth {
		p.requirePwd = "pass"
	}

	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if tlsCfg != nil {
		p.ln = tls.NewListener(raw, tlsCfg)
	} else {
		p.ln = raw
	}

	go p.acceptLoop()
	t.Cleanup(func() { _ = p.ln.Close() })
	return p
}

func (p *fakeProxy) acceptLoop() {
	for {
		c, err := p.ln.Accept()
		if err != nil {
			return
		}
		go p.handle(c)
	}
}

func (p *fakeProxy) handle(client net.Conn) {
	defer client.Close()

	br := bufio.NewReader(client)
	req, err := http.ReadRequest(br)
	if err != nil {
		return
	}

	p.mu.Lock()
	p.requests = append(p.requests, req.Method+" "+req.Host)
	p.gotAuth = append(p.gotAuth, req.Header.Get("Proxy-Authorization"))
	p.mu.Unlock()

	if req.Method != http.MethodConnect {
		_, _ = io.WriteString(client, "HTTP/1.1 400 Bad Request\r\n\r\n")
		return
	}

	if p.requirePwd != "" {
		auth := req.Header.Get("Proxy-Authorization")
		want := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:"+p.requirePwd))
		if auth != want {
			p.mu.Lock()
			p.rejected = append(p.rejected, req.Host)
			p.mu.Unlock()
			_, _ = io.WriteString(client, "HTTP/1.1 407 Proxy Authentication Required\r\n\r\n")
			return
		}
	}

	target, err := net.DialTimeout("tcp", req.Host, 5*time.Second)
	if err != nil {
		p.mu.Lock()
		p.rejected = append(p.rejected, req.Host+" (dial)")
		p.mu.Unlock()
		_, _ = io.WriteString(client, "HTTP/1.1 502 Bad Gateway\r\n\r\n")
		return
	}
	defer target.Close()

	p.mu.Lock()
	p.connected = append(p.connected, req.Host)
	p.mu.Unlock()

	_, _ = io.WriteString(client, "HTTP/1.1 200 Connection Established\r\n\r\n")

	// If the request parser buffered bytes past the CONNECT line, they belong to
	// the tunneled protocol and must be forwarded first.
	if br.Buffered() > 0 {
		buffered, _ := io.ReadAll(io.LimitReader(br, int64(br.Buffered())))
		_, _ = target.Write(buffered)
	}

	// Relay both directions until either side closes.
	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(target, br); done <- struct{}{} }()
	go func() { _, _ = io.Copy(client, target); done <- struct{}{} }()
	<-done
}

func (p *fakeProxy) addr() string { return p.ln.Addr().String() }

// spec builds the proxy URL as it would appear in configuration. The host and
// port are taken from the listener, so the loopback address is always correct.
func (p *fakeProxy) spec(scheme string, withAuth bool) string {
	host, port, _ := net.SplitHostPort(p.addr())
	cred := ""
	if withAuth {
		cred = "user:pass@"
	}
	if scheme == "" {
		// The bare form, which is what the project documents.
		return cred + host + ":" + port
	}
	return scheme + "://" + cred + host + ":" + port
}

func (p *fakeProxy) counts() (connected, rejected []string, auths []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.connected...),
		append([]string(nil), p.rejected...),
		append([]string(nil), p.gotAuth...)
}

// startEchoServer returns a TCP server that echoes, standing in for MongoDB's
// wire protocol. What matters is that a non-HTTP protocol survives the tunnel.
func startEchoServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.Copy(c, c)
			}(c)
		}
	}()
	return ln.Addr().String()
}

// TestDialerTunnelsRawProtocol is the core MongoDB case: arbitrary bytes, not an
// HTTP request, must arrive intact through the CONNECT tunnel.
func TestDialerTunnelsRawProtocol(t *testing.T) {
	origin := startEchoServer(t)
	p := newFakeProxy(t, false)

	ep, err := Parse(p.spec("", false))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ep.Dialer().DialContext(ctx, "tcp", origin)
	if err != nil {
		t.Fatalf("DialContext through proxy: %v", err)
	}
	defer conn.Close()

	payload := []byte("\x01\x02\x03 mongo wire bytes \xff\xfe")
	if _, err := conn.Write(payload); err != nil {
		t.Fatal(err)
	}

	got := make([]byte, len(payload))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("reading echo: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("tunnel corrupted data:\n got %q\nwant %q", got, payload)
	}

	connected, rejected, _ := p.counts()
	if len(connected) != 1 || connected[0] != origin {
		t.Errorf("proxy connected = %v, want [%s]", connected, origin)
	}
	if len(rejected) != 0 {
		t.Errorf("proxy rejected unexpectedly: %v", rejected)
	}
}

// TestDialerSendsProxyAuth verifies credentials reach the proxy, since without
// them a paid proxy answers 407 and the connection never establishes.
func TestDialerSendsProxyAuth(t *testing.T) {
	origin := startEchoServer(t)
	p := newFakeProxy(t, true)

	ep, err := Parse(p.spec("", true))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ep.Dialer().DialContext(ctx, "tcp", origin)
	if err != nil {
		t.Fatalf("DialContext with auth: %v", err)
	}
	defer conn.Close()

	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	_, _, auths := p.counts()
	if len(auths) == 0 {
		t.Fatal("proxy received no requests")
	}
	if auths[0] != want {
		t.Errorf("Proxy-Authorization = %q, want %q", auths[0], want)
	}
}

// TestDialerSurfacesAuthFailure ensures a wrong password is a clear error rather
// than a hang or a confusing protocol failure later.
func TestDialerSurfacesAuthFailure(t *testing.T) {
	origin := startEchoServer(t)
	p := newFakeProxy(t, true)

	// Correct host and port, wrong password.
	host, port, _ := net.SplitHostPort(p.addr())
	ep, err := Parse("user:WRONG@" + host + ":" + port)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = ep.Dialer().DialContext(ctx, "tcp", origin)
	if err == nil {
		t.Fatal("expected an error when the proxy rejects credentials")
	}
	if !strings.Contains(err.Error(), "407") {
		t.Errorf("error should mention the 407 status, got: %v", err)
	}
}

// TestDialerSurfacesUnreachableTarget verifies a target the proxy cannot reach
// produces an error naming the refused CONNECT.
func TestDialerSurfacesUnreachableTarget(t *testing.T) {
	p := newFakeProxy(t, false)

	// A port nothing listens on.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := ln.Addr().String()
	_ = ln.Close()

	ep, _ := Parse(p.spec("", false))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := ep.Dialer().DialContext(ctx, "tcp", closed); err == nil {
		t.Fatal("expected an error for an unreachable target")
	} else if !strings.Contains(err.Error(), "502") {
		t.Errorf("error should mention the 502 from the proxy, got: %v", err)
	}
}

// TestDialerRejectsNonTCP guards the guard: the tunnel is only meaningful for
// TCP, and silently accepting "unix" would produce a broken connection.
func TestDialerRejectsNonTCP(t *testing.T) {
	ep, _ := Parse("127.0.0.1:1")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := ep.Dialer().DialContext(ctx, "unix", "/tmp/x.sock"); err == nil {
		t.Error("expected an error when tunneling a non-TCP network")
	}
}

// TestDialerHonorsContextCancellation ensures a caller deadline aborts the
// connection attempt instead of blocking for the internal timeout.
func TestDialerHonorsContextCancellation(t *testing.T) {
	ep, _ := Parse("10.255.255.1:9") // unroutable, so the dial would hang

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := ep.Dialer().DialContext(ctx, "tcp", "example.com:27017")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error from a cancelled context")
	}
	if elapsed > 3*time.Second {
		t.Errorf("dial ignored the caller deadline, took %v", elapsed)
	}
	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context") {
		t.Logf("note: error was %v (not a bare DeadlineExceeded)", err)
	}
}

// TestDialerSatisfiesMongoContextDialer pins the contract mongo-driver requires:
// a type with DialContext(ctx, network, address) (net.Conn, error).
func TestDialerSatisfiesMongoContextDialer(t *testing.T) {
	type contextDialer interface {
		DialContext(ctx context.Context, network, address string) (net.Conn, error)
	}
	ep, _ := Parse("127.0.0.1:1")
	var _ contextDialer = ep.Dialer()
}

// TestHTTPTransportWorksThroughProxy exercises the Bot API path end-to-end: an
// HTTPS request must complete through the proxy.
func TestHTTPTransportWorksThroughProxy(t *testing.T) {
	origin := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "telegram-ok")
	}))
	defer origin.Close()

	p := newFakeProxy(t, false)
	ep, _ := Parse(p.spec("", false))

	tr := ep.HTTPTransport()
	// Trust the test server's self-signed certificate so the assertion is about
	// proxying, not certificate validation.
	tr.TLSClientConfig = origin.Client().Transport.(*http.Transport).TLSClientConfig

	client := &http.Client{Transport: tr, Timeout: 15 * time.Second}
	resp, err := client.Get(origin.URL)
	if err != nil {
		t.Fatalf("GET through proxy: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "telegram-ok" {
		t.Errorf("body = %q, want %q", body, "telegram-ok")
	}

	connected, _, _ := p.counts()
	if len(connected) == 0 {
		t.Error("proxy saw no CONNECT, so traffic did not go through it")
	}
}

// TestHTTPTransportSendsProxyAuth covers the authenticated case for the Bot API
// path, which is the configuration actually being deployed.
func TestHTTPTransportSendsProxyAuth(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer origin.Close()

	p := newFakeProxy(t, true)
	ep, _ := Parse(p.spec("", true))

	client := ep.HTTPClient()
	client.Timeout = 15 * time.Second
	resp, err := client.Get(origin.URL)
	if err != nil {
		t.Fatalf("GET through authenticated proxy: %v", err)
	}
	defer resp.Body.Close()

	_, _, auths := p.counts()
	if len(auths) == 0 || auths[0] == "" {
		t.Errorf("proxy got no Proxy-Authorization header: %v", auths)
	}
}

// --- https:// proxy (TLS to the proxy itself) ---------------------------

// trustProxyRoot returns an Endpoint configured to trust the proxy's own
// self-signed certificate. Production code leaves tlsConfig nil, which validates
// against the system root pool; a test proxy is never in that pool.
func trustProxyRoot(t *testing.T, spec string, cert tls.Certificate) *Endpoint {
	t.Helper()

	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(leaf)

	ep, err := Parse(spec)
	if err != nil {
		t.Fatal(err)
	}
	ep.tlsConfig = &tls.Config{
		RootCAs:    pool,
		ServerName: "127.0.0.1",
		MinVersion: tls.VersionTLS12,
	}
	return ep
}

// TestDialerThroughTLSProxy covers the https:// scheme, where a TLS handshake
// with the proxy must complete before the CONNECT request is sent.
func TestDialerThroughTLSProxy(t *testing.T) {
	origin := startEchoServer(t)
	p, cert := newTLSFakeProxy(t, false)

	ep := trustProxyRoot(t, p.spec("https", false), cert)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ep.Dialer().DialContext(ctx, "tcp", origin)
	if err != nil {
		t.Fatalf("DialContext through https proxy: %v", err)
	}
	defer conn.Close()

	payload := []byte("tls-proxy-tunnel-check")
	if _, err := conn.Write(payload); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, len(payload))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Errorf("got %q, want %q", got, payload)
	}

	connected, rejected, _ := p.counts()
	if len(connected) != 1 || connected[0] != origin {
		t.Errorf("proxy connected = %v, want [%s]", connected, origin)
	}
	if len(rejected) != 0 {
		t.Errorf("proxy rejected unexpectedly: %v", rejected)
	}
}

// TestTLSProxyFailureIsNotSilentlyIgnored is the security guard for wrapTLS: an
// untrusted proxy certificate must surface as an error. An earlier version of
// this code fell back to InsecureSkipVerify on failure, which would have
// disabled MITM protection with nothing in the logs to show it.
//
// This deliberately uses the production path (no injected tlsConfig), so the
// certificate is checked against the system root pool, which a self-signed test
// proxy is not in.
func TestTLSProxyFailureIsNotSilentlyIgnored(t *testing.T) {
	origin := startEchoServer(t)
	p, _ := newTLSFakeProxy(t, false)

	ep, err := Parse(p.spec("https", false))
	if err != nil {
		t.Fatal(err)
	}
	if ep.tlsConfig != nil {
		t.Fatal("expected the production path with no injected TLS config")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = ep.Dialer().DialContext(ctx, "tcp", origin)
	if err == nil {
		t.Fatal("expected a TLS verification failure, got a successful tunnel")
	}
	if !strings.Contains(err.Error(), "TLS handshake") {
		t.Errorf("error should name the TLS handshake, got: %v", err)
	}
}

// TestNoInsecureSkipVerifyInProductionPath asserts the production constructor
// never disables certificate validation.
func TestNoInsecureSkipVerifyInProductionPath(t *testing.T) {
	ep, err := Parse("https://user:pass@proxy.example:8443")
	if err != nil {
		t.Fatal(err)
	}
	if ep.tlsConfig != nil {
		t.Error("Parse must not pre-set a TLS config that could weaken validation")
	}
}

// TestTLSProxyRejectsHostnameMismatch proves verification is bound to the proxy
// host: a certificate that IS trusted but was issued for a different name must
// still be refused. Trusting the issuer alone would let any valid certificate
// from the pool stand in for the real proxy.
func TestTLSProxyRejectsHostnameMismatch(t *testing.T) {
	origin := startEchoServer(t)

	mismatched, err := selfSignedCertForOtherHost()
	if err != nil {
		t.Fatal(err)
	}
	p := startFakeProxy(t, false, &tls.Config{Certificates: []tls.Certificate{mismatched}})

	ep := trustProxyRoot(t, p.spec("https", false), mismatched)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := ep.Dialer().DialContext(ctx, "tcp", origin); err == nil {
		t.Fatal("a trusted cert for the wrong hostname must be rejected")
	} else if !strings.Contains(err.Error(), "TLS handshake") {
		t.Errorf("error should name the TLS handshake, got: %v", err)
	}
}

// selfSignedCert builds a certificate valid for 127.0.0.1, matching what the
// tests dial.
func selfSignedCert() (tls.Certificate, error) {
	return issueSelfSigned([]string{"localhost"}, []net.IP{net.ParseIP("127.0.0.1")})
}

// selfSignedCertForOtherHost builds a certificate that does NOT match
// 127.0.0.1, so a validating client must reject it.
func selfSignedCertForOtherHost() (tls.Certificate, error) {
	return issueSelfSigned([]string{"some-other-host.example"}, nil)
}

// issueSelfSigned creates a throwaway self-signed certificate for tests.
func issueSelfSigned(dnsNames []string, ips []net.IP) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              dnsNames,
		IPAddresses:           ips,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, nil
}
