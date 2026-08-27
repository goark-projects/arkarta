package tck

import (
	"context"
	"io"
	"reflect"
	"testing"

	"goark.dev/arkarta/websocket"
)

// SessionFactory 创建标准 WebSocket 会话。
type SessionFactory func(id string, connection websocket.Connection, options ...websocket.SessionOption) (*websocket.StandardSession, error)

// RunEndpointLifecycle 执行 WebSocket Endpoint 兼容性测试。
func RunEndpointLifecycle(t *testing.T, factory SessionFactory) {
	t.Helper()
	conn := newRecordingConnection(
		websocket.TextMessage("hello"),
		websocket.BinaryMessage([]byte{1, 2}),
		websocket.PongMessage([]byte("pong")),
		websocket.CloseMessage(websocket.NewCloseReason(websocket.CloseNormal, "bye")),
	)
	session, err := factory("s1", conn, websocket.WithSubprotocol("chat"), websocket.WithAttribute("tenant", "alpha"))
	if err != nil {
		t.Fatalf("factory failed: %v", err)
	}
	var events []string
	endpoint := websocket.EndpointFunc{
		Open: func(ctx context.Context, session websocket.Session) error {
			if session.Subprotocol() != "chat" {
				t.Fatalf("subprotocol = %q, want chat", session.Subprotocol())
			}
			if tenant, ok := session.Attribute("tenant"); !ok || tenant != "alpha" {
				t.Fatalf("tenant = %v/%v, want alpha/true", tenant, ok)
			}
			events = append(events, "open")
			return session.SendText(ctx, "ready")
		},
		Text: func(context.Context, websocket.Session, string) error {
			events = append(events, "text")
			return nil
		},
		Binary: func(context.Context, websocket.Session, []byte) error {
			events = append(events, "binary")
			return nil
		},
		Pong: func(context.Context, websocket.Session, []byte) error {
			events = append(events, "pong")
			return nil
		},
		Close: func(_ context.Context, _ websocket.Session, reason websocket.CloseReason) error {
			events = append(events, "close:"+reason.Reason())
			return nil
		},
	}
	if err := websocket.Serve(context.Background(), session, endpoint); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}
	want := []string{"open", "text", "binary", "pong", "close:bye"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if len(conn.writes) != 1 || conn.writes[0].Text() != "ready" {
		t.Fatalf("writes = %#v, want ready", conn.writes)
	}
}

type recordingConnection struct {
	reads  []websocket.Message
	writes []websocket.Message
}

func newRecordingConnection(reads ...websocket.Message) *recordingConnection {
	return &recordingConnection{reads: reads}
}

func (c *recordingConnection) Read(context.Context) (websocket.Message, error) {
	if len(c.reads) == 0 {
		return websocket.Message{}, io.EOF
	}
	message := c.reads[0]
	c.reads = c.reads[1:]
	return message, nil
}

func (c *recordingConnection) Write(_ context.Context, message websocket.Message) error {
	c.writes = append(c.writes, message)
	return nil
}

func (c *recordingConnection) Close(context.Context, websocket.CloseReason) error {
	return nil
}
