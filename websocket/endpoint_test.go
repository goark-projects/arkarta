package websocket

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestServeDispatchesEndpointLifecycle(t *testing.T) {
	t.Parallel()

	conn := newFakeConnection(
		TextMessage("hello"),
		BinaryMessage([]byte{1, 2}),
		PongMessage([]byte("pong")),
		CloseMessage(NewCloseReason(CloseNormal, "bye")),
	)
	session, err := NewSession("s1", conn, WithSubprotocol("chat"), WithAttribute("tenant", "alpha"))
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	var events []string
	endpoint := EndpointFunc{
		Open: func(ctx context.Context, session Session) error {
			if session.ID() != "s1" || session.Subprotocol() != "chat" {
				t.Fatalf("session id/subprotocol = %q/%q", session.ID(), session.Subprotocol())
			}
			if tenant, ok := session.Attribute("tenant"); !ok || tenant != "alpha" {
				t.Fatalf("tenant = %v/%v, want alpha/true", tenant, ok)
			}
			events = append(events, "open")
			return session.SendText(ctx, "ready")
		},
		Text: func(context.Context, Session, string) error {
			events = append(events, "text")
			return nil
		},
		Binary: func(context.Context, Session, []byte) error {
			events = append(events, "binary")
			return nil
		},
		Pong: func(context.Context, Session, []byte) error {
			events = append(events, "pong")
			return nil
		},
		Close: func(_ context.Context, _ Session, reason CloseReason) error {
			events = append(events, "close:"+reason.Reason())
			return nil
		},
	}

	if err := Serve(context.Background(), session, endpoint); err != nil {
		t.Fatalf("Serve failed: %v", err)
	}
	want := []string{"open", "text", "binary", "pong", "close:bye"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if len(conn.writes) != 1 || conn.writes[0].Text() != "ready" {
		t.Fatalf("writes = %#v, want ready", conn.writes)
	}
	if !session.Closed() {
		t.Fatal("session should be closed after close frame")
	}
}

func TestServeClosesOnEndpointError(t *testing.T) {
	t.Parallel()

	conn := newFakeConnection(TextMessage("boom"))
	session, err := NewSession("s1", conn)
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	cause := errors.New("handler failed")
	var sawError bool
	endpoint := EndpointFunc{
		Text: func(context.Context, Session, string) error {
			return cause
		},
		Error: func(_ context.Context, _ Session, err error) {
			sawError = errors.Is(err, cause)
		},
	}

	err = Serve(context.Background(), session, endpoint)
	if !errors.Is(err, cause) {
		t.Fatalf("Serve err = %v, want cause", err)
	}
	if !sawError {
		t.Fatal("endpoint should receive OnError")
	}
	if conn.closeReason.Code() != CloseUnexpectedCondition {
		t.Fatalf("close code = %d, want unexpected condition", conn.closeReason.Code())
	}
}

func TestJSONTextCodec(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}
	codec := JSONTextCodec[payload]{}
	text, err := codec.EncodeText(payload{Name: "arkarta"})
	if err != nil {
		t.Fatalf("EncodeText failed: %v", err)
	}
	got, err := codec.DecodeText(text)
	if err != nil {
		t.Fatalf("DecodeText failed: %v", err)
	}
	if got.Name != "arkarta" {
		t.Fatalf("payload = %#v, want arkarta", got)
	}
}

type fakeConnection struct {
	reads       []Message
	writes      []Message
	closeReason CloseReason
}

func newFakeConnection(reads ...Message) *fakeConnection {
	return &fakeConnection{reads: reads}
}

func (c *fakeConnection) Read(context.Context) (Message, error) {
	if len(c.reads) == 0 {
		return Message{}, io.EOF
	}
	message := c.reads[0]
	c.reads = c.reads[1:]
	return message, nil
}

func (c *fakeConnection) Write(_ context.Context, message Message) error {
	c.writes = append(c.writes, message)
	return nil
}

func (c *fakeConnection) Close(_ context.Context, reason CloseReason) error {
	c.closeReason = reason
	return nil
}
