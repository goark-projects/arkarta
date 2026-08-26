package servlet

import "strings"

// ConnectionInfo 描述一次请求所属连接的稳定元数据。
type ConnectionInfo struct {
	id                   string
	protocol             string
	protocolConnectionID string
	secure               bool
	localAddr            string
	remoteAddr           string
}

// ID 返回容器分配的连接 ID。
func (c ConnectionInfo) ID() string {
	return c.id
}

// Protocol 返回 HTTP 协议版本。
func (c ConnectionInfo) Protocol() string {
	return c.protocol
}

// ProtocolConnectionID 返回协议层连接 ID。
func (c ConnectionInfo) ProtocolConnectionID() string {
	return c.protocolConnectionID
}

// Secure 表示连接是否使用安全传输。
func (c ConnectionInfo) Secure() bool {
	return c.secure
}

// LocalAddr 返回本地网络地址。
func (c ConnectionInfo) LocalAddr() string {
	return c.localAddr
}

// RemoteAddr 返回远端网络地址。
func (c ConnectionInfo) RemoteAddr() string {
	return c.remoteAddr
}

// ConnectionInfo 返回当前请求的连接元数据。
func (r *Request) ConnectionInfo() ConnectionInfo {
	id := r.connectionID
	if id == "" {
		id = deriveConnectionID(r)
	}
	return ConnectionInfo{
		id:                   id,
		protocol:             r.Protocol(),
		protocolConnectionID: id,
		secure:               r.IsSecure(),
		localAddr:            r.LocalAddr(),
		remoteAddr:           r.RemoteAddr(),
	}
}

func deriveConnectionID(r *Request) string {
	parts := []string{r.LocalAddr(), r.RemoteAddr(), r.Protocol()}
	joined := strings.Join(parts, "|")
	if strings.Trim(joined, "|") == "" {
		return ""
	}
	return joined
}
