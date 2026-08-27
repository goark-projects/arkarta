package servlet_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goark.dev/arkarta/websocket"
	"goark.dev/arkarta/websocket/frame"
	servletws "goark.dev/arkarta/websocket/servlet"
)

func TestFrameConnectionReadsFragmentsAndAutoPongs(t *testing.T) {
	t.Parallel()

	conn := newDuplexUpgradeConn()
	writeClientFrame(t, conn.inbound, frame.New(frame.OpPing, []byte("p"), frame.WithMask(frame.MaskKey{1, 2, 3, 4})))
	writeClientFrame(t, conn.inbound, frame.New(frame.OpText, []byte("hel"), frame.WithFin(false), frame.WithMask(frame.MaskKey{4, 3, 2, 1})))
	writeClientFrame(t, conn.inbound, frame.New(frame.OpContinuation, []byte("lo"), frame.WithMask(frame.MaskKey{9, 8, 7, 6})))

	wsConn, err := servletws.NewFrameConnection(conn)
	if err != nil {
		t.Fatalf("NewFrameConnection failed: %v", err)
	}
	message, err := wsConn.Read(context.Background())
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if message.Type() != websocket.MessageText || message.Text() != "hello" {
		t.Fatalf("message = %v/%q, want text hello", message.Type(), message.Text())
	}

	pong, err := frame.Read(bytes.NewReader(conn.outbound.Bytes()), frame.WithMaskPolicy(frame.MaskForbidden))
	if err != nil {
		t.Fatalf("read pong failed: %v", err)
	}
	if pong.OpCode() != frame.OpPong || string(pong.Payload()) != "p" {
		t.Fatalf("pong = %v/%q, want pong p", pong.OpCode(), string(pong.Payload()))
	}
}

func TestFrameConnectionWritesServerFramesAndCloses(t *testing.T) {
	t.Parallel()

	conn := newDuplexUpgradeConn()
	wsConn, err := servletws.NewFrameConnection(conn)
	if err != nil {
		t.Fatalf("NewFrameConnection failed: %v", err)
	}
	if err := wsConn.Write(context.Background(), websocket.TextMessage("ready")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := wsConn.Close(context.Background(), websocket.NewCloseReason(websocket.CloseNormal, "bye")); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	reader := bytes.NewReader(conn.outbound.Bytes())
	text, err := frame.Read(reader, frame.WithMaskPolicy(frame.MaskForbidden))
	if err != nil {
		t.Fatalf("read text failed: %v", err)
	}
	closeFrame, err := frame.Read(reader, frame.WithMaskPolicy(frame.MaskForbidden))
	if err != nil {
		t.Fatalf("read close failed: %v", err)
	}
	if text.OpCode() != frame.OpText || string(text.Payload()) != "ready" {
		t.Fatalf("text = %v/%q", text.OpCode(), string(text.Payload()))
	}
	code, reason, err := frame.ParseClosePayload(closeFrame.Payload())
	if err != nil {
		t.Fatalf("ParseClosePayload failed: %v", err)
	}
	if closeFrame.OpCode() != frame.OpClose || code != 1000 || reason != "bye" || !conn.closed {
		t.Fatalf("close = opcode %v code %d reason %q closed %v", closeFrame.OpCode(), code, reason, conn.closed)
	}
}

func TestServeEndpointUsesFrameConnection(t *testing.T) {
	t.Parallel()

	conn := newDuplexUpgradeConn()
	closePayload, err := frame.ClosePayload(uint16(websocket.CloseNormal), "bye")
	if err != nil {
		t.Fatalf("ClosePayload failed: %v", err)
	}
	writeClientFrame(t, conn.inbound, frame.New(frame.OpText, []byte("hello"), frame.WithMask(frame.MaskKey{1, 1, 1, 1})))
	writeClientFrame(t, conn.inbound, frame.New(frame.OpClose, closePayload, frame.WithMask(frame.MaskKey{2, 2, 2, 2})))

	var events []string
	endpoint := websocket.EndpointFunc{
		Open: func(ctx context.Context, session websocket.Session) error {
			if session.ID() != "s1" {
				t.Fatalf("session id = %q, want s1", session.ID())
			}
			if session.Subprotocol() != "chat" {
				t.Fatalf("subprotocol = %q, want chat", session.Subprotocol())
			}
			events = append(events, "open")
			return session.SendText(ctx, "ready")
		},
		Text: func(_ context.Context, _ websocket.Session, text string) error {
			events = append(events, "text:"+text)
			return nil
		},
		Close: func(_ context.Context, _ websocket.Session, reason websocket.CloseReason) error {
			events = append(events, "close:"+reason.Reason())
			return nil
		},
	}
	handshake := newHandshakeWithSubprotocol(t, "chat")
	if err := servletws.ServeEndpoint(context.Background(), "s1", handshake, conn, endpoint); err != nil {
		t.Fatalf("ServeEndpoint failed: %v", err)
	}
	if want := []string{"open", "text:hello", "close:bye"}; !reflectDeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func writeClientFrame(t *testing.T, buffer *bytes.Buffer, next frame.Frame) {
	t.Helper()
	if err := frame.Write(buffer, next); err != nil {
		t.Fatalf("Write client frame failed: %v", err)
	}
}

type duplexUpgradeConn struct {
	inbound  *bytes.Buffer
	outbound bytes.Buffer
	closed   bool
}

func newDuplexUpgradeConn() *duplexUpgradeConn {
	return &duplexUpgradeConn{inbound: &bytes.Buffer{}}
}

func (c *duplexUpgradeConn) Read(data []byte) (int, error) {
	return c.inbound.Read(data)
}

func (c *duplexUpgradeConn) Write(data []byte) (int, error) {
	return c.outbound.Write(data)
}

func (c *duplexUpgradeConn) Close() error {
	c.closed = true
	return nil
}

func (c *duplexUpgradeConn) LocalAddr() net.Addr {
	return testAddr("local")
}

func (c *duplexUpgradeConn) RemoteAddr() net.Addr {
	return testAddr("remote")
}

func (c *duplexUpgradeConn) SetDeadline(time.Time) error {
	return nil
}

func (c *duplexUpgradeConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *duplexUpgradeConn) SetWriteDeadline(time.Time) error {
	return nil
}

var _ io.Reader = (*duplexUpgradeConn)(nil)

func newHandshakeWithSubprotocol(t *testing.T, subprotocol string) websocket.Handshake {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/ws", nil)
	request.Header.Set("Connection", "Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	request.Header.Set("Sec-WebSocket-Version", websocket.ProtocolVersion)
	request.Header.Set("Sec-WebSocket-Protocol", subprotocol)
	handshake, err := websocket.NewHandshaker(websocket.WithSubprotocols(subprotocol)).Accept(request)
	if err != nil {
		t.Fatalf("Accept failed: %v", err)
	}
	return handshake
}

func reflectDeepEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
