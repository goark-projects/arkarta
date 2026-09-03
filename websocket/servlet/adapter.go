package servlet

import (
	"context"
	"errors"
	"net/http"

	arkservlet "goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/upgrade"
	"goark.dev/arkarta/websocket"
)

// Adapter 将 WebSocket 标准握手桥接到 Servlet Upgrade Profile。
type Adapter struct {
	handshaker *websocket.Handshaker
}

// NewAdapter 创建 Servlet WebSocket 升级适配器。
func NewAdapter(options ...websocket.HandshakeOption) *Adapter {
	return &Adapter{handshaker: websocket.NewHandshaker(options...)}
}

// Upgrade 使用临时适配器执行 WebSocket 握手和 Servlet 协议升级。
func Upgrade(ctx context.Context, req *arkservlet.Request, res arkservlet.Response, handler Handler, options ...websocket.HandshakeOption) (websocket.Handshake, error) {
	return NewAdapter(options...).Upgrade(ctx, req, res, handler)
}

// Upgrade 校验握手、写出 101 响应，并把升级后的连接移交给处理器。
func (a *Adapter) Upgrade(ctx context.Context, req *arkservlet.Request, res arkservlet.Response, handler Handler) (websocket.Handshake, error) {
	if handler == nil {
		return websocket.Handshake{}, ErrNilHandler
	}
	if req == nil {
		return websocket.Handshake{}, arkservlet.ErrNilRequestInput
	}
	if res == nil {
		return websocket.Handshake{}, arkservlet.ErrNilResponse
	}
	handshaker := websocket.NewHandshaker()
	if a != nil && a.handshaker != nil {
		handshaker = a.handshaker
	}
	handshake, err := handshaker.AcceptRequest(req.Method(), req.Header())
	if err != nil {
		writeHandshakeError(res, err)
		return websocket.Handshake{}, err
	}
	err = upgrade.HTTP(ctx, req, res, upgrade.HandlerFunc(func(ctx context.Context, conn upgrade.Connection) error {
		if err := WriteHandshakeResponse(conn, handshake); err != nil {
			return err
		}
		return handler.ServeWebSocket(ctx, handshake, conn)
	}))
	return handshake, err
}

func writeHandshakeError(res arkservlet.Response, err error) {
	if res == nil || res.Committed() {
		return
	}
	statusCode := http.StatusBadRequest
	message := http.StatusText(statusCode)
	var handshakeErr *websocket.HandshakeError
	if errors.As(err, &handshakeErr) {
		statusCode = handshakeErr.StatusCode()
		message = handshakeErr.PublicMessage()
		if errors.Is(err, websocket.ErrUnsupportedVersion) {
			res.Header().Set("Sec-WebSocket-Version", websocket.ProtocolVersion)
		}
	}
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.SetStatus(statusCode)
	_, _ = res.WriteString(message + "\n")
}
