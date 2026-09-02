package servlet_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	arkservlet "goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/upgrade"
	"goark.dev/arkarta/websocket"
	servletws "goark.dev/arkarta/websocket/servlet"
)

func TestUpgradeWritesHandshakeResponseAndDelegates(t *testing.T) {
	t.Parallel()

	req := newUpgradeRequest(t)
	res := newUpgradeResponse()
	var seenSubprotocol string

	handshake, err := servletws.Upgrade(context.Background(), req, res, servletws.HandlerFunc(func(_ context.Context, handshake websocket.Handshake, _ upgrade.Connection) error {
		seenSubprotocol = handshake.Subprotocol()
		return nil
	}), websocket.WithSubprotocols("chat"))
	if err != nil {
		t.Fatalf("Upgrade failed: %v", err)
	}
	if handshake.Subprotocol() != "chat" || seenSubprotocol != "chat" {
		t.Fatalf("subprotocol = %q/%q, want chat", handshake.Subprotocol(), seenSubprotocol)
	}
	raw := strings.ToLower(res.conn.String())
	if !strings.HasPrefix(raw, "http/1.1 101 switching protocols\r\n") {
		t.Fatalf("response = %q, want HTTP 101", res.conn.String())
	}
	if !strings.Contains(raw, "sec-websocket-accept: s3pplmbitxaq9kygzzhzrbk+xoo=") ||
		!strings.Contains(raw, "sec-websocket-protocol: chat") {
		t.Fatalf("response headers = %q, want accept and subprotocol", res.conn.String())
	}
	if !res.upgraded || !res.Committed() {
		t.Fatal("response should be upgraded and committed")
	}
}

func TestUpgradeRejectsInvalidHandshakeBeforeConnectionHandoff(t *testing.T) {
	t.Parallel()

	req := newUpgradeRequest(t)
	req.HTTPRequest().Header.Set("Sec-WebSocket-Key", "bad")
	res := newUpgradeResponse()

	_, err := servletws.Upgrade(context.Background(), req, res, servletws.HandlerFunc(func(context.Context, websocket.Handshake, upgrade.Connection) error {
		t.Fatal("handler must not run for invalid handshake")
		return nil
	}))
	if !errors.Is(err, websocket.ErrInvalidHandshake) {
		t.Fatalf("Upgrade err = %v, want ErrInvalidHandshake", err)
	}
	if res.Status() != http.StatusBadRequest || !strings.Contains(res.body.String(), "invalid websocket key") {
		t.Fatalf("error response = status %d body %q", res.Status(), res.body.String())
	}
	if res.upgraded {
		t.Fatal("invalid handshake must not upgrade connection")
	}
}

func newUpgradeRequest(t *testing.T) *arkservlet.Request {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/ws", nil)
	request.Header.Set("Connection", "keep-alive, Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	request.Header.Set("Sec-WebSocket-Version", websocket.ProtocolVersion)
	request.Header.Set("Sec-WebSocket-Protocol", "superchat, chat")
	req, err := arkservlet.NewRequest(request)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	return req
}

type upgradeResponse struct {
	header    arkservlet.Header
	status    int
	committed bool
	body      bytes.Buffer
	conn      *recordingConn
	upgraded  bool
}

func newUpgradeResponse() *upgradeResponse {
	return &upgradeResponse{
		header: arkservlet.NewHeader(),
		status: http.StatusOK,
		conn:   &recordingConn{},
	}
}

func (r *upgradeResponse) Header() arkservlet.Header {
	return r.header
}

func (r *upgradeResponse) SetStatus(code int) {
	if !r.committed {
		r.status = code
	}
}

func (r *upgradeResponse) Status() int {
	return r.status
}

func (r *upgradeResponse) Write(data []byte) (int, error) {
	r.committed = true
	return r.body.Write(data)
}

func (r *upgradeResponse) WriteString(value string) (int, error) {
	return r.Write([]byte(value))
}

func (r *upgradeResponse) Flush() error {
	r.committed = true
	return nil
}

func (r *upgradeResponse) Committed() bool {
	return r.committed
}

func (r *upgradeResponse) Reset() error {
	if r.committed {
		return arkservlet.ErrResponseCommitted
	}
	r.header = arkservlet.NewHeader()
	r.status = http.StatusOK
	r.body.Reset()
	return nil
}

func (r *upgradeResponse) BodyWriter() io.Writer {
	return &r.body
}

func (r *upgradeResponse) UpgradeHTTP(ctx context.Context, _ *arkservlet.Request, handler upgrade.Handler) error {
	r.upgraded = true
	r.committed = true
	return handler.ServeUpgrade(ctx, r.conn)
}

type recordingConn struct {
	bytes.Buffer
}

func (c *recordingConn) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (c *recordingConn) Close() error {
	return nil
}

func (c *recordingConn) LocalAddr() net.Addr {
	return testAddr("local")
}

func (c *recordingConn) RemoteAddr() net.Addr {
	return testAddr("remote")
}

func (c *recordingConn) SetDeadline(time.Time) error {
	return nil
}

func (c *recordingConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *recordingConn) SetWriteDeadline(time.Time) error {
	return nil
}

type testAddr string

func (a testAddr) Network() string {
	return string(a)
}

func (a testAddr) String() string {
	return string(a)
}
