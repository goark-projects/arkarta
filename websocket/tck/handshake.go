package tck

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goark.dev/arkarta/websocket"
)

// HandshakerFactory 创建 WebSocket HTTP 握手协商器。
type HandshakerFactory func(options ...websocket.HandshakeOption) *websocket.Handshaker

// RunHandshake 执行 WebSocket HTTP 握手兼容性测试。
func RunHandshake(t *testing.T, factory HandshakerFactory) {
	t.Helper()
	t.Run("accepts_valid_handshake", func(t *testing.T) {
		request := NewHandshakeRequest()
		handshake, err := factory(websocket.WithSubprotocols("chat")).Accept(request)
		if err != nil {
			t.Fatalf("Accept failed: %v", err)
		}
		if handshake.AcceptValue() != "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" || handshake.Subprotocol() != "chat" {
			t.Fatalf("handshake accept/subprotocol = %q/%q", handshake.AcceptValue(), handshake.Subprotocol())
		}
	})
	t.Run("selects_subprotocol_by_client_order", func(t *testing.T) {
		request := NewHandshakeRequest()
		handshake, err := factory(websocket.WithSubprotocols("chat", "superchat")).Accept(request)
		if err != nil {
			t.Fatalf("Accept failed: %v", err)
		}
		if handshake.Subprotocol() != "superchat" {
			t.Fatalf("subprotocol = %q, want client preferred superchat", handshake.Subprotocol())
		}
	})
	t.Run("negotiates_permessage_deflate", func(t *testing.T) {
		request := NewHandshakeRequest()
		request.Header.Set("Sec-WebSocket-Extensions", "permessage-deflate; client_max_window_bits")
		handshake, err := factory(websocket.WithExtensions(websocket.NewPerMessageDeflate(
			websocket.WithServerNoContextTakeover(true),
		))).Accept(request)
		if err != nil {
			t.Fatalf("Accept failed: %v", err)
		}
		header := websocket.FormatExtensions(handshake.Extensions())
		if !strings.Contains(header, websocket.ExtensionPerMessageDeflate) || !strings.Contains(header, "server_no_context_takeover") {
			t.Fatalf("extensions = %q, want permessage-deflate", header)
		}
	})
	t.Run("writes_http_switching_protocols", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		_, err := factory().AcceptHTTP(recorder, NewHandshakeRequest())
		if err != nil {
			t.Fatalf("AcceptHTTP failed: %v", err)
		}
		if recorder.Code != http.StatusSwitchingProtocols || recorder.Header().Get("Sec-WebSocket-Accept") == "" {
			t.Fatalf("status/accept = %d/%q, want 101/accept", recorder.Code, recorder.Header().Get("Sec-WebSocket-Accept"))
		}
	})
	t.Run("rejects_bad_key", func(t *testing.T) {
		request := NewHandshakeRequest()
		request.Header.Set("Sec-WebSocket-Key", "bad")
		_, err := factory().Accept(request)
		if !errors.Is(err, websocket.ErrInvalidHandshake) {
			t.Fatalf("Accept err = %v, want ErrInvalidHandshake", err)
		}
	})
}

// NewHandshakeRequest 创建标准 TCK 握手请求。
func NewHandshakeRequest() *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/ws", nil)
	request.Header.Set("Connection", "keep-alive, Upgrade")
	request.Header.Set("Upgrade", "websocket")
	request.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	request.Header.Set("Sec-WebSocket-Version", websocket.ProtocolVersion)
	request.Header.Set("Sec-WebSocket-Protocol", "superchat, chat")
	return request
}
