package upgrade

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"goark.dev/arkarta/servlet"
)

func TestHTTPDelegatesToContainerUpgrader(t *testing.T) {
	t.Parallel()

	req := mustUpgradeRequest(t)
	res := &fakeUpgradeResponse{plainResponse: newPlainResponse(), conn: fakeConn{}}
	called := false
	err := HTTP(context.Background(), req, res, HandlerFunc(func(_ context.Context, conn Connection) error {
		called = true
		_, ok := conn.(fakeConn)
		if !ok {
			t.Fatalf("conn = %T, want fakeConn", conn)
		}
		return nil
	}))
	if err != nil {
		t.Fatalf("HTTP failed: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}

func TestHTTPRejectsUnsupportedResponse(t *testing.T) {
	t.Parallel()

	req := mustUpgradeRequest(t)
	err := HTTP(context.Background(), req, newPlainResponse(), HandlerFunc(func(context.Context, Connection) error {
		return nil
	}))
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("HTTP err = %v, want ErrUnsupported", err)
	}
}

func TestSwitchingProtocolsSetsHeaders(t *testing.T) {
	t.Parallel()

	res := newPlainResponse()
	if err := SwitchingProtocols(res, "websocket"); err != nil {
		t.Fatalf("SwitchingProtocols failed: %v", err)
	}
	if res.Status() != http.StatusSwitchingProtocols ||
		res.Header().Get("Connection") != "Upgrade" ||
		res.Header().Get("Upgrade") != "websocket" {
		t.Fatalf("response = status %d headers %#v", res.Status(), res.Header())
	}
}

func mustUpgradeRequest(t *testing.T) *servlet.Request {
	t.Helper()
	req, err := servlet.NewRequest((&http.Request{}).WithContext(context.Background()))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	return req
}

type fakeUpgradeResponse struct {
	*plainResponse
	conn Connection
}

func (r *fakeUpgradeResponse) UpgradeHTTP(ctx context.Context, _ *servlet.Request, handler Handler) error {
	return handler.ServeUpgrade(ctx, r.conn)
}

type plainResponse struct {
	header    servlet.Header
	status    int
	committed bool
}

func newPlainResponse() *plainResponse {
	return &plainResponse{header: servlet.NewHeader(), status: http.StatusOK}
}

func (r *plainResponse) Header() servlet.Header {
	return r.header
}

func (r *plainResponse) SetStatus(code int) {
	r.status = code
}

func (r *plainResponse) Status() int {
	return r.status
}

func (r *plainResponse) Write(data []byte) (int, error) {
	r.committed = true
	return len(data), nil
}

func (r *plainResponse) WriteString(value string) (int, error) {
	r.committed = true
	return len(value), nil
}

func (r *plainResponse) Flush() error {
	r.committed = true
	return nil
}

func (r *plainResponse) Committed() bool {
	return r.committed
}

func (r *plainResponse) Reset() error {
	return nil
}

func (r *plainResponse) BodyWriter() io.Writer {
	return io.Discard
}

type fakeConn struct {
	net.Conn
}

func (fakeConn) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (fakeConn) Write(data []byte) (int, error) {
	return len(data), nil
}

func (fakeConn) Close() error {
	return nil
}

func (fakeConn) LocalAddr() net.Addr {
	return fakeAddr("local")
}

func (fakeConn) RemoteAddr() net.Addr {
	return fakeAddr("remote")
}

func (fakeConn) SetDeadline(time.Time) error {
	return nil
}

func (fakeConn) SetReadDeadline(time.Time) error {
	return nil
}

func (fakeConn) SetWriteDeadline(time.Time) error {
	return nil
}

type fakeAddr string

func (a fakeAddr) Network() string {
	return string(a)
}

func (a fakeAddr) String() string {
	return string(a)
}
