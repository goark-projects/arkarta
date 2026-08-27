package websocket

// HandshakeOption 定制 WebSocket HTTP 握手协商器。
type HandshakeOption func(*Handshaker)

// WithSubprotocols 设置服务端支持的子协议，协商时遵循客户端偏好顺序。
func WithSubprotocols(protocols ...string) HandshakeOption {
	return func(handshaker *Handshaker) {
		handshaker.subprotocols = appendTokens(handshaker.subprotocols, protocols...)
	}
}

// WithExtensions 设置服务端支持的 WebSocket 扩展协商器。
func WithExtensions(extensions ...ExtensionNegotiator) HandshakeOption {
	return func(handshaker *Handshaker) {
		for _, extension := range extensions {
			if extension != nil {
				handshaker.extensions = append(handshaker.extensions, extension)
			}
		}
	}
}
