package servlet

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// ErrNoCookie 表示请求中不存在指定 Cookie。
var ErrNoCookie = errors.New("arkarta/servlet: named cookie not present")

// SameSite 表示 Cookie SameSite 策略。
type SameSite uint8

const (
	SameSiteDefaultMode SameSite = iota
	SameSiteLaxMode
	SameSiteStrictMode
	SameSiteNoneMode
)

// Cookie 表示与容器实现无关的 HTTP Cookie。
type Cookie struct {
	Name        string
	Value       string
	Path        string
	Domain      string
	Expires     time.Time
	MaxAge      int
	Secure      bool
	HTTPOnly    bool
	SameSite    SameSite
	Partitioned bool
}

// String 按 Set-Cookie 字段格式序列化 Cookie。
func (c *Cookie) String() string {
	if c == nil || !validCookieName(c.Name) {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(c.Name) + len(c.Value) + len(c.Path) + len(c.Domain) + 64)
	builder.WriteString(c.Name)
	builder.WriteByte('=')
	builder.WriteString(sanitizeCookieValue(c.Value))
	if validCookiePath(c.Path) {
		builder.WriteString("; Path=")
		builder.WriteString(c.Path)
	}
	if domain := sanitizeCookieDomain(c.Domain); domain != "" {
		builder.WriteString("; Domain=")
		builder.WriteString(domain)
	}
	if validCookieExpires(c.Expires) {
		builder.WriteString("; Expires=")
		builder.WriteString(c.Expires.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	}
	switch {
	case c.MaxAge < 0:
		builder.WriteString("; Max-Age=0")
	case c.MaxAge > 0:
		builder.WriteString("; Max-Age=")
		builder.WriteString(strconv.Itoa(c.MaxAge))
	}
	if c.HTTPOnly {
		builder.WriteString("; HttpOnly")
	}
	if c.Secure {
		builder.WriteString("; Secure")
	}
	switch c.SameSite {
	case SameSiteLaxMode:
		builder.WriteString("; SameSite=Lax")
	case SameSiteStrictMode:
		builder.WriteString("; SameSite=Strict")
	case SameSiteNoneMode:
		builder.WriteString("; SameSite=None")
	}
	if c.Partitioned {
		builder.WriteString("; Partitioned")
	}
	return builder.String()
}

func parseRequestCookies(header Header) []*Cookie {
	if header == nil {
		return nil
	}
	var cookies []*Cookie
	for _, line := range header.Values("Cookie") {
		for _, pair := range strings.Split(line, ";") {
			name, value, ok := strings.Cut(strings.TrimSpace(pair), "=")
			if !ok || !validCookieName(name) || strings.HasPrefix(name, "$") {
				continue
			}
			value = strings.TrimSpace(value)
			if len(value) > 1 && value[0] == '"' {
				unquoted, err := strconv.Unquote(value)
				if err != nil {
					continue
				}
				value = unquoted
			}
			cookies = append(cookies, &Cookie{Name: name, Value: value})
		}
	}
	return cookies
}

func validCookieName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c <= 0x20 || c >= 0x7f || strings.ContainsRune("()<>@,;:\\\"/[]?={}", rune(c)) {
			return false
		}
	}
	return true
}

func sanitizeCookieValue(value string) string {
	return strings.Map(func(r rune) rune {
		if r == 0x21 || r >= 0x23 && r <= 0x2b || r >= 0x2d && r <= 0x3a || r >= 0x3c && r <= 0x5b || r >= 0x5d && r <= 0x7e {
			return r
		}
		return -1
	}, value)
}

func validCookiePath(path string) bool {
	return path != "" && !strings.ContainsAny(path, ";\r\n")
}

func sanitizeCookieDomain(domain string) string {
	domain = strings.TrimPrefix(strings.TrimSpace(domain), ".")
	if domain == "" || strings.ContainsAny(domain, ";\r\n ") {
		return ""
	}
	return domain
}

func validCookieExpires(expires time.Time) bool {
	return !expires.IsZero() && expires.Year() >= 1601
}
