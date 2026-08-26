package session

import (
	"net/url"
	"strings"
	"unicode"

	"goark.dev/arkarta/servlet"
)

const (
	// DefaultURLParameterName 是 Servlet URL 重写的默认路径参数名。
	DefaultURLParameterName = "jsessionid"
)

// URLRewriteOption 定制 URL 重写行为。
type URLRewriteOption func(*URLRewriter) error

// URLRewriter 实现 Servlet encodeURL / encodeRedirectURL 语义。
type URLRewriter struct {
	parameterName   string
	cookieName      string
	cookiePreferred bool
}

// NewURLRewriter 创建 URL 重写器。
func NewURLRewriter(options ...URLRewriteOption) (*URLRewriter, error) {
	rewriter := &URLRewriter{
		parameterName:   DefaultURLParameterName,
		cookieName:      DefaultCookieName,
		cookiePreferred: true,
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(rewriter); err != nil {
			return nil, err
		}
	}
	return rewriter, nil
}

// WithRewriteParameterName 设置 URL 重写路径参数名。
func WithRewriteParameterName(name string) URLRewriteOption {
	return func(rewriter *URLRewriter) error {
		if !validPathParameterName(name) {
			return ErrInvalidURLRewriteConfig
		}
		rewriter.parameterName = name
		return nil
	}
}

// WithRewriteCookieName 设置 cookie 优先判断使用的 Cookie 名称。
func WithRewriteCookieName(name string) URLRewriteOption {
	return func(rewriter *URLRewriter) error {
		if name == "" {
			return ErrInvalidCookieConfig
		}
		rewriter.cookieName = name
		return nil
	}
}

// WithCookiePreferred 设置是否在请求已携带会话 Cookie 时跳过 URL 重写。
func WithCookiePreferred(preferred bool) URLRewriteOption {
	return func(rewriter *URLRewriter) error {
		rewriter.cookiePreferred = preferred
		return nil
	}
}

// EncodeURL 使用默认重写器编码普通 URL。
func EncodeURL(req *servlet.Request, rawURL, sessionID string) (string, error) {
	return defaultURLRewriter.EncodeURL(req, rawURL, sessionID)
}

// EncodeRedirectURL 使用默认重写器编码重定向 URL。
func EncodeRedirectURL(req *servlet.Request, rawURL, sessionID string) (string, error) {
	return defaultURLRewriter.EncodeRedirectURL(req, rawURL, sessionID)
}

// EncodeURL 在 URL 路径中写入会话 ID。
func (r *URLRewriter) EncodeURL(req *servlet.Request, rawURL, sessionID string) (string, error) {
	if r == nil {
		return rawURL, ErrInvalidURLRewriteConfig
	}
	if rawURL == "" || sessionID == "" || r.hasSessionCookie(req) {
		return rawURL, nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if shouldSkipRewrite(req, parsed) || parsed.Path == "" {
		return rawURL, nil
	}
	parsed.Path = rewritePathParameter(parsed.Path, r.parameterName, sessionID)
	parsed.RawPath = ""
	return parsed.String(), nil
}

// EncodeRedirectURL 在重定向 URL 路径中写入会话 ID。
func (r *URLRewriter) EncodeRedirectURL(req *servlet.Request, rawURL, sessionID string) (string, error) {
	return r.EncodeURL(req, rawURL, sessionID)
}

func (r *URLRewriter) hasSessionCookie(req *servlet.Request) bool {
	if !r.cookiePreferred || req == nil {
		return false
	}
	cookie, err := req.Cookie(r.cookieName)
	return err == nil && cookie.Value != ""
}

func shouldSkipRewrite(req *servlet.Request, parsed *url.URL) bool {
	if parsed == nil {
		return true
	}
	if parsed.Scheme != "" && parsed.Host == "" {
		return true
	}
	if parsed.Host == "" || req == nil {
		return false
	}
	return !strings.EqualFold(parsed.Host, req.Host())
}

func rewritePathParameter(path, name, value string) string {
	clean := stripPathParameter(path, name)
	index := strings.LastIndex(clean, "/")
	if index < 0 {
		return clean + ";" + name + "=" + value
	}
	return clean[:index+1] + clean[index+1:] + ";" + name + "=" + value
}

func stripPathParameter(path, name string) string {
	segments := strings.Split(path, "/")
	prefix := ";" + name + "="
	for i, segment := range segments {
		parts := strings.Split(segment, ";")
		if len(parts) == 1 {
			continue
		}
		dst := parts[:1]
		for _, part := range parts[1:] {
			if !strings.HasPrefix(part, prefix[1:]) {
				dst = append(dst, part)
			}
		}
		segments[i] = strings.Join(dst, ";")
	}
	return strings.Join(segments, "/")
}

func validPathParameterName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' {
			continue
		}
		return false
	}
	return true
}

var defaultURLRewriter = &URLRewriter{
	parameterName:   DefaultURLParameterName,
	cookieName:      DefaultCookieName,
	cookiePreferred: true,
}
