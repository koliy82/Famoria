// Package proxy parses HTTP(S) proxy settings and builds the two kinds of
// transport the bot needs: an *http.Transport for HTTPS-based APIs, and a raw
// CONNECT tunnel for protocols that are not HTTP (MongoDB's wire protocol).
package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// connectTimeout bounds establishing the TCP connection to the proxy and the
// CONNECT exchange. It deliberately does not bound the tunneled traffic that
// follows, which can be a long-lived MongoDB connection.
const connectTimeout = 30 * time.Second

// Endpoint is a parsed, normalized HTTP(S) proxy.
type Endpoint struct {
	url *url.URL

	// tlsConfig overrides the TLS settings used when the proxy itself is reached
	// over TLS (an https:// proxy URL). It is nil in production, meaning the
	// proxy's certificate is validated against the system root pool; tests in
	// this package set it to trust a self-signed certificate.
	tlsConfig *tls.Config
}

// Parse normalizes a proxy specification.
//
// Accepted forms:
//
//	user:pass@host:port          (defaults to http, the common CONNECT proxy)
//	http://user:pass@host:port
//	https://user:pass@host:port   (TLS to the proxy itself)
//
// The scheme controls how the proxy is reached, not what it may carry: both an
// http:// and an https:// proxy can tunnel HTTPS traffic via CONNECT. A bare
// "user:pass@host:port" defaults to http:// because that is what proxy providers
// serve; write https:// explicitly to talk TLS to the proxy.
//
// Credentials with special characters must be percent-encoded, since the string
// is parsed as a URL.
func Parse(raw string) (*Endpoint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("proxy: empty proxy URL")
	}

	// A bare "user:pass@host:port" has no scheme, and url.Parse would then treat
	// "user:pass@host" as scheme:opaque.
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("proxy: invalid URL %q: %w", MaskURL(raw), err)
	}

	switch u.Scheme {
	case "http", "https":
	default:
		return nil, fmt.Errorf("proxy: unsupported scheme %q (want http or https)", u.Scheme)
	}

	host := u.Hostname()
	port := u.Port()
	if host == "" || port == "" {
		return nil, fmt.Errorf("proxy: %q must include a host and a port", MaskURL(raw))
	}

	return &Endpoint{url: u}, nil
}

// MustParse is Parse for configuration paths where an unusable proxy means the
// service cannot start at all, so failing fast beats silently connecting direct.
func MustParse(raw string) *Endpoint {
	ep, err := Parse(raw)
	if err != nil {
		panic(err)
	}
	return ep
}

// URL returns the normalized proxy URL, credentials included. Callers must not
// log it; use Masked.
func (e *Endpoint) URL() *url.URL {
	cp := *e.url
	return &cp
}

// Masked returns the proxy URL with the password removed, safe for logs.
func (e *Endpoint) Masked() string {
	return MaskURL(e.url.String())
}

// hasAuth reports whether the proxy requires credentials.
func (e *Endpoint) hasAuth() bool { return e.url.User != nil }

// MaskURL redacts the password from a proxy URL string.
//
// It applies the same scheme normalization as Parse before masking, because the
// documented configuration form "user:pass@host:port" has no scheme: url.Parse
// would read "user" as the scheme and the rest as an opaque value, find no
// userinfo, and hand the password straight back. As a last resort anything still
// shaped like credentials is replaced, so a secret never reaches a log or an
// error message through this function.
func MaskURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw
	}

	candidate := trimmed
	if !strings.Contains(candidate, "://") {
		candidate = "http://" + candidate
	}

	if u, err := url.Parse(candidate); err == nil && u.User != nil {
		cp := *u
		cp.User = url.User(u.User.Username())
		// Drop the normalization scheme if the input had none, so the masked
		// string still resembles what the operator wrote.
		masked := cp.String()
		if !strings.Contains(trimmed, "://") {
			masked = strings.TrimPrefix(masked, cp.Scheme+"://")
		}
		return masked
	}

	// Unparsable, but still redact anything that looks like user:pass@ before the
	// first host separator.
	if at := strings.LastIndex(trimmed, "@"); at > 0 {
		if colon := strings.Index(trimmed[:at], ":"); colon >= 0 {
			return trimmed[:colon+1] + "*****" + trimmed[at:]
		}
	}
	return raw
}

// HTTPTransport returns a transport that routes requests through the proxy.
//
// It starts from the default transport's clone so dial timeouts and keep-alive
// stay as they normally are, with two deliberate changes:
//
//   - MaxIdleConnsPerHost is raised. Go's default of 2 is far too small here:
//     the bot talks to a single host (api.telegram.org) from long polling,
//     message sends, media uploads and the Telegram log sink at the same time.
//     At 2 idle connections the rest are torn down and re-established through
//     the proxy, adding a CONNECT round-trip to nearly every request.
//   - Response-level timeouts are NOT set. Telegram long polling holds a request
//     open for tens of seconds before answering, so a ResponseHeaderTimeout
//     would abort every poll.
func (e *Endpoint) HTTPTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.Proxy = http.ProxyURL(e.url)
	t.MaxIdleConnsPerHost = maxIdleConnsPerHost
	return t
}

// maxIdleConnsPerHost is sized for one API host with several concurrent callers.
const maxIdleConnsPerHost = 32

// HTTPClient returns an http.Client routed through the proxy, ready to hand to
// telego's WithHTTPClient.
func (e *Endpoint) HTTPClient() *http.Client {
	return &http.Client{Transport: e.HTTPTransport()}
}

// Dialer is a net.Dialer-shaped function, satisfying mongo-driver's
// options.ContextDialer.
type Dialer func(ctx context.Context, network, address string) (net.Conn, error)

// DialContext implements options.ContextDialer.
func (d Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d(ctx, network, address)
}

// Dialer returns a dialer that reaches the target by opening an HTTP CONNECT
// tunnel through the proxy.
//
// This exists because mongo-driver v1 has no proxy option of its own: the
// MongoDB wire protocol is not HTTP, so it cannot go through http.Transport.
// The returned connection is the raw tunnel — the driver still applies its own
// TLS for mongodb+srv or tls=true URIs on top of it.
func (e *Endpoint) Dialer() Dialer {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" && network != "tcp4" && network != "tcp6" {
			return nil, fmt.Errorf("proxy: cannot tunnel network %q", network)
		}
		return e.connect(ctx, address)
	}
}

// connect dials the proxy, performs the CONNECT exchange, and returns the
// tunnel ready for the caller's protocol.
func (e *Endpoint) connect(ctx context.Context, address string) (net.Conn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	var d net.Dialer
	rawConn, err := d.DialContext(dialCtx, "tcp", net.JoinHostPort(e.url.Hostname(), e.url.Port()))
	if err != nil {
		return nil, fmt.Errorf("proxy: dial %s: %w", e.Masked(), err)
	}

	// Every failure path below must close the socket that was actually opened, so
	// keep it separate from the possibly-TLS-wrapped connection.
	conn := rawConn
	defer func() {
		if err != nil {
			_ = rawConn.Close()
		}
	}()

	// Reach an https:// proxy over TLS before tunneling through it.
	if e.url.Scheme == "https" {
		conn, err = e.wrapTLS(dialCtx, rawConn)
		if err != nil {
			return nil, err
		}
	}

	// Bound the CONNECT exchange itself, then lift the deadline so the tunnel is
	// not killed mid-use: a MongoDB connection lives far longer than 30s.
	if dl, ok := dialCtx.Deadline(); ok {
		if derr := conn.SetDeadline(dl); derr != nil {
			err = fmt.Errorf("proxy: set deadline: %w", derr)
			return nil, err
		}
	}
	if err = e.tunnel(conn, address); err != nil {
		return nil, err
	}
	if err = conn.SetDeadline(time.Time{}); err != nil {
		err = fmt.Errorf("proxy: clear deadline: %w", err)
		return nil, err
	}
	return conn, nil
}

// wrapTLS upgrades the connection to the proxy with TLS, so a CONNECT request
// can be sent to an https:// proxy.
//
// The certificate is validated against the proxy host. Failures are reported as
// errors rather than worked around: silently falling back to
// InsecureSkipVerify would disable MITM protection with nothing in the logs to
// show it happened.
func (e *Endpoint) wrapTLS(ctx context.Context, conn net.Conn) (net.Conn, error) {
	cfg := e.tlsConfig
	if cfg == nil {
		cfg = &tls.Config{
			ServerName: e.url.Hostname(),
			MinVersion: tls.VersionTLS12,
		}
	} else {
		// Clone so a caller-supplied config is never mutated here, and keep SNI
		// correct for the proxy host unless the caller already pinned it.
		cp := cfg.Clone()
		if cp.ServerName == "" {
			cp.ServerName = e.url.Hostname()
		}
		cfg = cp
	}

	tlsConn := tls.Client(conn, cfg)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return nil, fmt.Errorf("proxy: TLS handshake with %s: %w", e.Masked(), err)
	}
	return tlsConn, nil
}

// tunnel writes the CONNECT request and validates the proxy's answer.
func (e *Endpoint) tunnel(conn net.Conn, address string) error {
	req := &http.Request{
		Method: http.MethodConnect,
		// Opaque is required: ReadResponse rejects a CONNECT request whose URL
		// parses as a normal path, and the target here is a bare host:port.
		URL:    &url.URL{Opaque: address},
		Host:   address,
		Header: make(http.Header),
		Proto:  "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
	}
	if e.hasAuth() {
		user := e.url.User.Username()
		pass, _ := e.url.User.Password()
		req.Header.Set("Proxy-Authorization", "Basic "+
			base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))
	}

	if err := req.Write(conn); err != nil {
		return fmt.Errorf("proxy: write CONNECT: %w", err)
	}

	// A bufio.Reader is needed to parse the response without consuming bytes that
	// belong to the tunneled protocol.
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		return fmt.Errorf("proxy: read CONNECT response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy: CONNECT %s refused: %s", address, resp.Status)
	}

	// Any bytes the proxy sent right after the response line belong to the
	// tunneled protocol and must not be discarded.
	if br.Buffered() > 0 {
		return errors.New("proxy: unexpected data after CONNECT response")
	}
	return nil
}
