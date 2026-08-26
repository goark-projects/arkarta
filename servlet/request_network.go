package servlet

import (
	"net"
	"net/http"
	"strconv"
	"strings"
)

// ServerName 返回请求目标主机名。
func (r *Request) ServerName() string {
	host, _, _ := splitHostPortDefault(r.Host(), r.Scheme())
	return host
}

// ServerPort 返回请求目标端口。
func (r *Request) ServerPort() int {
	_, port, ok := splitHostPortDefault(r.Host(), r.Scheme())
	if !ok {
		return 0
	}
	return port
}

// RemoteHost 返回远端主机名或 IP；标准实现不做反向 DNS 查询。
func (r *Request) RemoteHost() string {
	host, _, ok := splitAddr(r.RemoteAddr())
	if !ok {
		return ""
	}
	return host
}

// RemotePort 返回远端端口。
func (r *Request) RemotePort() int {
	_, port, ok := splitAddr(r.RemoteAddr())
	if !ok {
		return 0
	}
	return port
}

// LocalAddr 返回当前连接的本地网络地址。
func (r *Request) LocalAddr() string {
	addr, ok := r.httpRequest.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok || addr == nil {
		return ""
	}
	return addr.String()
}

// LocalName 返回当前连接的本地主机名或 IP。
func (r *Request) LocalName() string {
	host, _, ok := splitAddr(r.LocalAddr())
	if !ok {
		return ""
	}
	return host
}

// LocalPort 返回当前连接的本地端口。
func (r *Request) LocalPort() int {
	_, port, ok := splitAddr(r.LocalAddr())
	if !ok {
		return 0
	}
	return port
}

func splitHostPortDefault(value, scheme string) (string, int, bool) {
	host, port, ok := splitAddr(value)
	if ok {
		return host, port, true
	}
	host = strings.Trim(value, "[]")
	if host == "" {
		return "", 0, false
	}
	switch strings.ToLower(scheme) {
	case "https":
		return host, 443, true
	case "http", "":
		return host, 80, true
	default:
		return host, 0, true
	}
}

func splitAddr(value string) (string, int, bool) {
	if value == "" {
		return "", 0, false
	}
	host, portText, err := net.SplitHostPort(value)
	if err != nil {
		return strings.Trim(value, "[]"), 0, false
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return host, 0, false
	}
	return strings.Trim(host, "[]"), port, true
}
