package session

import (
	"net/http"

	"goark.dev/arkarta/servlet"
)

const (
	// DefaultCookieName 是 Servlet 生态默认的会话 Cookie 名称。
	DefaultCookieName = "JSESSIONID"
)

// CookieConfig 描述会话 Cookie 的写出策略。
type CookieConfig struct {
	name     string
	path     string
	domain   string
	maxAge   int
	secure   bool
	httpOnly bool
	sameSite http.SameSite
}

func defaultCookieConfig() CookieConfig {
	return CookieConfig{
		name:     DefaultCookieName,
		httpOnly: true,
		sameSite: http.SameSiteLaxMode,
	}
}

// Name 返回 Cookie 名称。
func (c CookieConfig) Name() string {
	return c.name
}

// Path 返回固定 Cookie Path；空值表示按请求上下文路径计算。
func (c CookieConfig) Path() string {
	return c.path
}

// Domain 返回 Cookie Domain。
func (c CookieConfig) Domain() string {
	return c.domain
}

// MaxAge 返回 Cookie Max-Age。
func (c CookieConfig) MaxAge() int {
	return c.maxAge
}

// Secure 表示是否强制写出 Secure。
func (c CookieConfig) Secure() bool {
	return c.secure
}

// HTTPOnly 表示是否写出 HttpOnly。
func (c CookieConfig) HTTPOnly() bool {
	return c.httpOnly
}

// SameSite 返回 Cookie SameSite 策略。
func (c CookieConfig) SameSite() http.SameSite {
	return c.sameSite
}

func (c CookieConfig) cookie(req *servlet.Request, id string) *http.Cookie {
	path := c.path
	if path == "" && req != nil {
		path = req.ContextPath()
	}
	if path == "" {
		path = "/"
	}
	return &http.Cookie{
		Name:     c.name,
		Value:    id,
		Path:     path,
		Domain:   c.domain,
		MaxAge:   c.maxAge,
		Secure:   c.secure || (req != nil && req.IsSecure()),
		HttpOnly: c.httpOnly,
		SameSite: c.sameSite,
	}
}
