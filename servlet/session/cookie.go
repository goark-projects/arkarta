package session

import (
	"strings"

	"goark.dev/arkarta/servlet"
)

const (
	// DefaultCookieName 是 Servlet 生态默认的会话 Cookie 名称。
	DefaultCookieName = "JSESSIONID"
	// AttributeCookieConfig 保存 WebApp 级 Session Cookie 配置。
	AttributeCookieConfig = "arkarta.servlet.session.cookie_config"
)

// CookieConfig 描述会话 Cookie 的写出策略。
type CookieConfig struct {
	name     string
	path     string
	domain   string
	maxAge   int
	secure   bool
	httpOnly bool
	sameSite servlet.SameSite
}

// CookieConfigOption 定制 Session Cookie 配置。
type CookieConfigOption func(*CookieConfig) error

// NewCookieConfig 创建 Session Cookie 配置。
func NewCookieConfig(options ...CookieConfigOption) (CookieConfig, error) {
	config := defaultCookieConfig()
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&config); err != nil {
			return CookieConfig{}, err
		}
	}
	return config, nil
}

// WithCookieConfigName 设置 Session Cookie 名称。
func WithCookieConfigName(name string) CookieConfigOption {
	return func(config *CookieConfig) error {
		if !validCookieName(name) {
			return ErrInvalidCookieConfig
		}
		config.name = name
		return nil
	}
}

// WithCookieConfigPath 设置 Session Cookie Path。
func WithCookieConfigPath(path string) CookieConfigOption {
	return func(config *CookieConfig) error {
		config.path = path
		return nil
	}
}

// WithCookieConfigDomain 设置 Session Cookie Domain。
func WithCookieConfigDomain(domain string) CookieConfigOption {
	return func(config *CookieConfig) error {
		config.domain = domain
		return nil
	}
}

// WithCookieConfigMaxAge 设置 Session Cookie Max-Age。
func WithCookieConfigMaxAge(maxAge int) CookieConfigOption {
	return func(config *CookieConfig) error {
		config.maxAge = maxAge
		return nil
	}
}

// WithCookieConfigSecure 设置是否强制写出 Secure。
func WithCookieConfigSecure(secure bool) CookieConfigOption {
	return func(config *CookieConfig) error {
		config.secure = secure
		return nil
	}
}

// WithCookieConfigHTTPOnly 设置是否写出 HttpOnly。
func WithCookieConfigHTTPOnly(httpOnly bool) CookieConfigOption {
	return func(config *CookieConfig) error {
		config.httpOnly = httpOnly
		return nil
	}
}

// WithCookieConfigSameSite 设置 SameSite 策略。
func WithCookieConfigSameSite(sameSite servlet.SameSite) CookieConfigOption {
	return func(config *CookieConfig) error {
		config.sameSite = sameSite
		return nil
	}
}

func defaultCookieConfig() CookieConfig {
	return CookieConfig{
		name:     DefaultCookieName,
		httpOnly: true,
		sameSite: servlet.SameSiteLaxMode,
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
func (c CookieConfig) SameSite() servlet.SameSite {
	return c.sameSite
}

func (c CookieConfig) cookie(req *servlet.Request, id string) *servlet.Cookie {
	path := c.path
	if path == "" && req != nil {
		path = req.ContextPath()
	}
	if path == "" {
		path = "/"
	}
	return &servlet.Cookie{
		Name:     c.name,
		Value:    id,
		Path:     path,
		Domain:   c.domain,
		MaxAge:   c.maxAge,
		Secure:   c.secure || (req != nil && req.IsSecure()),
		HTTPOnly: c.httpOnly,
		SameSite: c.sameSite,
	}
}

// ConfigureCookie 将 Session Cookie 配置绑定到 WebApp。
func ConfigureCookie(app *servlet.WebApp, config CookieConfig) error {
	if app == nil || !validCookieName(config.name) {
		return ErrInvalidCookieConfig
	}
	switch app.State() {
	case servlet.WebAppStateNew, servlet.WebAppStateInitialized:
		app.SetAttribute(AttributeCookieConfig, config)
		return nil
	default:
		return ErrCookieConfigLocked
	}
}

// CookieConfigFor 返回 WebApp 级 Session Cookie 配置。
func CookieConfigFor(app *servlet.WebApp) (CookieConfig, bool) {
	if app == nil {
		return CookieConfig{}, false
	}
	value, ok := app.Attribute(AttributeCookieConfig)
	if !ok {
		return CookieConfig{}, false
	}
	config, ok := value.(CookieConfig)
	return config, ok
}

func validCookieName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	for _, r := range name {
		if r <= 0x20 || r >= 0x7f {
			return false
		}
		switch r {
		case '(', ')', '<', '>', '@', ',', ';', ':', '\\', '"', '/', '[', ']', '?', '=', '{', '}':
			return false
		}
	}
	return true
}
