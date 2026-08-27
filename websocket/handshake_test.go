package websocket

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandshakerAcceptsValidRequest(t *testing.T) {
	t.Parallel()

	request := newHandshakeRequest()
	handshake, err := NewHandshaker(WithSubprotocols("chat")).Accept(request)
	if err != nil {
		t.Fatalf("Accept failed: %v", err)
	}
	if handshake.AcceptValue() != "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" {
		t.Fatalf("accept = %q, want RFC sample", handshake.AcceptValue())
	}
	if handshake.Subprotocol() != "chat" {
		t.Fatalf("subprotocol = %q, want chat", handshake.Subprotocol())
	}
}

func TestHandshakerSelectsSubprotocolByClientOrder(t *testing.T) {
	t.Parallel()

	request := newHandshakeRequest()
	handshake, err := NewHandshaker(WithSubprotocols("chat", "superchat")).Accept(request)
	if err != nil {
		t.Fatalf("Accept failed: %v", err)
	}
	if handshake.Subprotocol() != "superchat" {
		t.Fatalf("subprotocol = %q, want client preferred superchat", handshake.Subprotocol())
	}
}

func TestHandshakerNegotiatesExtensionsAndWritesHTTP(t *testing.T) {
	t.Parallel()

	request := newHandshakeRequest()
	request.Header.Set("Sec-WebSocket-Extensions", "permessage-deflate; client_max_window_bits")
	recorder := httptest.NewRecorder()
	handshake, err := NewHandshaker(
		WithExtensions(NewPerMessageDeflate(
			WithServerNoContextTakeover(true),
			WithClientNoContextTakeover(true),
		)),
	).AcceptHTTP(recorder, request)
	if err != nil {
		t.Fatalf("AcceptHTTP failed: %v", err)
	}

	if recorder.Code != http.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want 101", recorder.Code)
	}
	if recorder.Header().Get("Sec-WebSocket-Accept") != handshake.AcceptValue() {
		t.Fatalf("response accept = %q, want handshake value", recorder.Header().Get("Sec-WebSocket-Accept"))
	}
	extension := recorder.Header().Get("Sec-WebSocket-Extensions")
	if !strings.Contains(extension, ExtensionPerMessageDeflate) ||
		!strings.Contains(extension, "client_no_context_takeover") ||
		!strings.Contains(extension, "server_no_context_takeover") {
		t.Fatalf("extensions = %q, want permessage-deflate no-context", extension)
	}
}

func TestExtensionParametersAreNormalized(t *testing.T) {
	t.Parallel()

	extension, ok := NewExtension("X-Test", map[string]string{
		"Client_No_Context_Takeover": "",
		"bad;name":                   "ignored",
	})
	if !ok {
		t.Fatal("NewExtension failed")
	}
	if extension.Name() != "x-test" {
		t.Fatalf("name = %q, want x-test", extension.Name())
	}
	if _, ok := extension.Parameter("client_no_context_takeover"); !ok {
		t.Fatal("normalized parameter missing")
	}
	if _, ok := extension.Parameter("bad;name"); ok {
		t.Fatal("invalid parameter should be ignored")
	}
}

func TestHandshakerRejectsInvalidRequests(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/ws", nil)
	_, err := NewHandshaker().Accept(request)
	var handshakeErr *HandshakeError
	if !errors.As(err, &handshakeErr) || !errors.Is(err, ErrInvalidHandshake) || handshakeErr.StatusCode() != http.StatusMethodNotAllowed {
		t.Fatalf("POST err = %v, want 405 invalid handshake", err)
	}

	request = newHandshakeRequest()
	request.Header.Set("Sec-WebSocket-Version", "12")
	recorder := httptest.NewRecorder()
	_, err = NewHandshaker().AcceptHTTP(recorder, request)
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("version err = %v, want ErrUnsupportedVersion", err)
	}
	if recorder.Code != http.StatusUpgradeRequired || recorder.Header().Get("Sec-WebSocket-Version") != ProtocolVersion {
		t.Fatalf("status/version = %d/%q, want 426/13", recorder.Code, recorder.Header().Get("Sec-WebSocket-Version"))
	}
}

func TestPerMessageDeflateRoundTripAndLimit(t *testing.T) {
	t.Parallel()

	extension := NewPerMessageDeflate()
	compressed, err := extension.CompressMessage([]byte(strings.Repeat("arkarta", 16)))
	if err != nil {
		t.Fatalf("CompressMessage failed: %v", err)
	}
	if strings.HasSuffix(string(compressed), "\x00\x00\xff\xff") {
		t.Fatalf("compressed payload keeps permessage-deflate tail: %x", compressed)
	}
	data, err := extension.DecompressMessage(compressed, 1024)
	if err != nil {
		t.Fatalf("DecompressMessage failed: %v", err)
	}
	if string(data) != strings.Repeat("arkarta", 16) {
		t.Fatalf("data = %q, want original", string(data))
	}
	if _, err := extension.DecompressMessage(compressed, 8); !errors.Is(err, ErrMessageTooLarge) {
		t.Fatalf("limited DecompressMessage err = %v, want ErrMessageTooLarge", err)
	}
}

func newHandshakeRequest() *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/ws", nil)
	request.Header.Set("Connection", "keep-alive, Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	request.Header.Set("Sec-WebSocket-Version", ProtocolVersion)
	request.Header.Set("Sec-WebSocket-Protocol", "superchat, chat")
	return request
}
