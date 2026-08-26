package session

import "net/http"

// AccessorOption 定制会话请求访问器。
type AccessorOption func(*Accessor) error

// WithCookieName 设置会话 Cookie 名称。
func WithCookieName(name string) AccessorOption {
	return func(accessor *Accessor) error {
		if name == "" {
			return ErrInvalidCookieConfig
		}
		accessor.cookie.name = name
		return nil
	}
}

// WithCookiePath 设置固定 Cookie Path。
func WithCookiePath(path string) AccessorOption {
	return func(accessor *Accessor) error {
		accessor.cookie.path = path
		return nil
	}
}

// WithCookieDomain 设置 Cookie Domain。
func WithCookieDomain(domain string) AccessorOption {
	return func(accessor *Accessor) error {
		accessor.cookie.domain = domain
		return nil
	}
}

// WithCookieMaxAge 设置 Cookie Max-Age。
func WithCookieMaxAge(maxAge int) AccessorOption {
	return func(accessor *Accessor) error {
		accessor.cookie.maxAge = maxAge
		return nil
	}
}

// WithCookieSecure 设置是否强制写出 Secure。
func WithCookieSecure(secure bool) AccessorOption {
	return func(accessor *Accessor) error {
		accessor.cookie.secure = secure
		return nil
	}
}

// WithCookieHTTPOnly 设置是否写出 HttpOnly。
func WithCookieHTTPOnly(httpOnly bool) AccessorOption {
	return func(accessor *Accessor) error {
		accessor.cookie.httpOnly = httpOnly
		return nil
	}
}

// WithCookieSameSite 设置 SameSite 策略。
func WithCookieSameSite(sameSite http.SameSite) AccessorOption {
	return func(accessor *Accessor) error {
		accessor.cookie.sameSite = sameSite
		return nil
	}
}

// WithTrackingModes 设置允许的会话跟踪模式。
func WithTrackingModes(modes ...TrackingMode) AccessorOption {
	return func(accessor *Accessor) error {
		policy, err := NewTrackingPolicy(modes...)
		if err != nil {
			return err
		}
		accessor.tracking = policy
		return nil
	}
}

// WithTrackingPolicy 设置会话跟踪策略。
func WithTrackingPolicy(policy TrackingPolicy) AccessorOption {
	return func(accessor *Accessor) error {
		if len(policy.Modes()) == 0 {
			accessor.tracking = DefaultTrackingPolicy()
			return nil
		}
		accessor.tracking = policy
		return nil
	}
}
