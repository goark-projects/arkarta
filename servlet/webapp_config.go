package servlet

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	// DefaultVirtualServerName 是标准上下文默认虚拟主机名。
	DefaultVirtualServerName = "default"
	// DefaultCharacterEncoding 是请求与响应的默认字符编码。
	DefaultCharacterEncoding = "utf-8"
)

// DefaultSessionTimeout 是标准会话默认空闲超时。
const DefaultSessionTimeout = 30 * time.Minute

// ErrInvalidWebAppConfig 表示 WebApp 配置非法。
var ErrInvalidWebAppConfig = errors.New("arkarta/servlet: invalid web app config")

// WithVirtualServerName 设置虚拟主机名。
func WithVirtualServerName(name string) WebAppOption {
	return func(app *WebApp) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return ErrInvalidWebAppConfig
		}
		app.virtualServerName = name
		return nil
	}
}

// WithRequestCharacterEncoding 设置默认请求字符编码。
func WithRequestCharacterEncoding(charset string) WebAppOption {
	return func(app *WebApp) error {
		charset = strings.TrimSpace(charset)
		if charset == "" {
			return ErrInvalidWebAppConfig
		}
		app.requestCharacterEncoding = charset
		return nil
	}
}

// WithResponseCharacterEncoding 设置默认响应字符编码。
func WithResponseCharacterEncoding(charset string) WebAppOption {
	return func(app *WebApp) error {
		charset = strings.TrimSpace(charset)
		if charset == "" {
			return ErrInvalidWebAppConfig
		}
		app.responseCharacterEncoding = charset
		return nil
	}
}

// WithSessionTimeout 设置默认会话空闲超时。
func WithSessionTimeout(timeout time.Duration) WebAppOption {
	return func(app *WebApp) error {
		if timeout < 0 {
			return fmt.Errorf("%w: negative session timeout", ErrInvalidWebAppConfig)
		}
		app.sessionTimeout = timeout
		return nil
	}
}

// VirtualServerName 返回虚拟主机名。
func (a *WebApp) VirtualServerName() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.virtualServerName
}

// RequestCharacterEncoding 返回默认请求字符编码。
func (a *WebApp) RequestCharacterEncoding() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.requestCharacterEncoding
}

// ResponseCharacterEncoding 返回默认响应字符编码。
func (a *WebApp) ResponseCharacterEncoding() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.responseCharacterEncoding
}

// SessionTimeout 返回默认会话空闲超时。
func (a *WebApp) SessionTimeout() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessionTimeout
}
