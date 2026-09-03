package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

const (
	// ProtocolVersion 是 Arkarta WebSocket 当前支持的 RFC 6455 协议版本。
	ProtocolVersion = "13"

	handshakeGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
)

// Handshaker 负责校验 HTTP Upgrade 请求并生成 WebSocket 握手响应。
type Handshaker struct {
	subprotocols []string
	extensions   []ExtensionNegotiator
}

// HandshakeHeader 提供 WebSocket 握手所需的最小请求头读取契约。
type HandshakeHeader interface {
	Get(name string) string
	Values(name string) []string
}

// NewHandshaker 创建 WebSocket HTTP 握手协商器。
func NewHandshaker(options ...HandshakeOption) *Handshaker {
	handshaker := &Handshaker{}
	for _, option := range options {
		if option != nil {
			option(handshaker)
		}
	}
	return handshaker
}

// Handshake 描述一次成功协商后的握手结果。
type Handshake struct {
	accept      string
	subprotocol string
	extensions  []Extension
	header      http.Header
}

// Accept 校验请求并返回握手响应。
func (h *Handshaker) Accept(request *http.Request) (Handshake, error) {
	if request == nil {
		return Handshake{}, ErrNilHTTPRequest
	}
	return h.AcceptRequest(request.Method, request.Header)
}

// AcceptRequest 校验传输中立的请求方法与请求头并返回握手响应。
func (h *Handshaker) AcceptRequest(method string, header HandshakeHeader) (Handshake, error) {
	if header == nil {
		return Handshake{}, ErrNilHTTPRequest
	}
	if method != http.MethodGet {
		return Handshake{}, newHandshakeError(http.StatusMethodNotAllowed, "websocket handshake requires GET", ErrInvalidHandshake)
	}
	if !headerContainsToken(header, "Connection", "Upgrade") {
		return Handshake{}, newHandshakeError(http.StatusBadRequest, "missing Connection upgrade token", ErrInvalidHandshake)
	}
	if !headerContainsToken(header, "Upgrade", "websocket") {
		return Handshake{}, newHandshakeError(http.StatusBadRequest, "missing websocket upgrade header", ErrInvalidHandshake)
	}
	if header.Get("Sec-WebSocket-Version") != ProtocolVersion {
		return Handshake{}, newHandshakeError(http.StatusUpgradeRequired, "unsupported websocket version", ErrUnsupportedVersion)
	}
	key := strings.TrimSpace(header.Get("Sec-WebSocket-Key"))
	if !validHandshakeKey(key) {
		return Handshake{}, newHandshakeError(http.StatusBadRequest, "invalid websocket key", ErrInvalidHandshake)
	}

	subprotocol := selectSubprotocol(headerTokens(header.Values("Sec-WebSocket-Protocol")), h.subprotocols)
	extensions := negotiateExtensions(ParseExtensions(header.Values("Sec-WebSocket-Extensions")...), h.extensions)
	responseHeader := http.Header{}
	responseHeader.Set("Upgrade", "websocket")
	responseHeader.Set("Connection", "Upgrade")
	responseHeader.Set("Sec-WebSocket-Accept", acceptKey(key))
	if subprotocol != "" {
		responseHeader.Set("Sec-WebSocket-Protocol", subprotocol)
	}
	if len(extensions) > 0 {
		responseHeader.Set("Sec-WebSocket-Extensions", FormatExtensions(extensions))
	}
	return Handshake{
		accept:      responseHeader.Get("Sec-WebSocket-Accept"),
		subprotocol: subprotocol,
		extensions:  cloneExtensions(extensions),
		header:      cloneHeader(responseHeader),
	}, nil
}

// AcceptHTTP 执行握手并写出标准 HTTP 101 响应头。
func (h *Handshaker) AcceptHTTP(writer http.ResponseWriter, request *http.Request) (Handshake, error) {
	if writer == nil {
		return Handshake{}, ErrNilResponseWriter
	}
	handshake, err := h.Accept(request)
	if err != nil {
		WriteHandshakeError(writer, err)
		return Handshake{}, err
	}
	return handshake, handshake.WriteTo(writer)
}

// AcceptValue 返回 Sec-WebSocket-Accept 响应值。
func (h Handshake) AcceptValue() string {
	return h.accept
}

// Subprotocol 返回协商后的子协议。
func (h Handshake) Subprotocol() string {
	return h.subprotocol
}

// Extensions 返回协商后的扩展列表。
func (h Handshake) Extensions() []Extension {
	return cloneExtensions(h.extensions)
}

// Header 返回握手响应头副本。
func (h Handshake) Header() http.Header {
	return cloneHeader(h.header)
}

// WriteTo 写出标准 HTTP 101 握手响应。
func (h Handshake) WriteTo(writer http.ResponseWriter) error {
	if writer == nil {
		return ErrNilResponseWriter
	}
	for name, values := range h.header {
		for _, value := range values {
			writer.Header().Add(name, value)
		}
	}
	writer.WriteHeader(http.StatusSwitchingProtocols)
	return nil
}

// WriteHandshakeError 按握手错误状态写出安全错误响应。
func WriteHandshakeError(writer http.ResponseWriter, err error) {
	if writer == nil {
		return
	}
	statusCode := http.StatusBadRequest
	message := http.StatusText(statusCode)
	var handshakeErr *HandshakeError
	if errors.As(err, &handshakeErr) {
		statusCode = handshakeErr.StatusCode()
		message = handshakeErr.PublicMessage()
		if errors.Is(err, ErrUnsupportedVersion) {
			writer.Header().Set("Sec-WebSocket-Version", ProtocolVersion)
		}
	}
	http.Error(writer, message, statusCode)
}

func validHandshakeKey(key string) bool {
	data, err := base64.StdEncoding.DecodeString(key)
	return err == nil && len(data) == 16
}

func acceptKey(key string) string {
	sum := sha1.Sum([]byte(key + handshakeGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func headerContainsToken(header HandshakeHeader, name, token string) bool {
	for _, candidate := range headerTokens(header.Values(name)) {
		if strings.EqualFold(candidate, token) {
			return true
		}
	}
	return false
}

func selectSubprotocol(clientProtocols, serverProtocols []string) string {
	if len(clientProtocols) == 0 || len(serverProtocols) == 0 {
		return ""
	}
	supported := make(map[string]string, len(serverProtocols))
	for _, protocol := range serverProtocols {
		supported[strings.ToLower(protocol)] = protocol
	}
	for _, protocol := range clientProtocols {
		if selected, ok := supported[strings.ToLower(protocol)]; ok {
			return selected
		}
	}
	return ""
}

func negotiateExtensions(offers []Extension, negotiators []ExtensionNegotiator) []Extension {
	if len(offers) == 0 || len(negotiators) == 0 {
		return nil
	}
	result := make([]Extension, 0, len(negotiators))
	for _, negotiator := range negotiators {
		if extension, ok := negotiator.Negotiate(offers); ok {
			result = append(result, extension)
		}
	}
	return result
}

func cloneHeader(src http.Header) http.Header {
	dst := make(http.Header, len(src))
	for name, values := range src {
		dst[name] = append([]string(nil), values...)
	}
	return dst
}
