package servlet

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

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
		if r == 0x21 ||
			r >= 0x23 && r <= 0x2b || r >= 0x2d && r <= 0x3a ||
			r >= 0x3c && r <= 0x5b || r >= 0x5d && r <= 0x7e {
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

// Locale 表示经过标准化的语言标签。
type Locale struct {
	tag      string
	language string
	region   string
}

// NewLocale 从 BCP 47 风格标签构造 Locale。
func NewLocale(tag string) (Locale, bool) {
	tag = strings.TrimSpace(tag)
	if tag == "" || strings.ContainsAny(tag, " \t\r\n") {
		return Locale{}, false
	}
	parts := strings.Split(strings.ReplaceAll(tag, "_", "-"), "-")
	if parts[0] == "" {
		return Locale{}, false
	}
	language := strings.ToLower(parts[0])
	region := ""
	if len(parts) > 1 && parts[1] != "" {
		region = strings.ToUpper(parts[1])
	}
	normalized := language
	if region != "" {
		normalized += "-" + region
	}
	return Locale{tag: normalized, language: language, region: region}, true
}

// String 返回标准化语言标签。
func (l Locale) String() string {
	return l.tag
}

// Tag 返回标准化语言标签。
func (l Locale) Tag() string {
	return l.tag
}

// Language 返回语言子标签。
func (l Locale) Language() string {
	return l.language
}

// Region 返回地区子标签。
func (l Locale) Region() string {
	return l.region
}

// Locale 返回 Accept-Language 中优先级最高的语言。
func (r *Request) Locale() (Locale, bool) {
	locales := r.Locales()
	if len(locales) == 0 {
		return Locale{}, false
	}
	return locales[0], true
}

// Locales 按客户端声明优先级返回语言列表。
func (r *Request) Locales() []Locale {
	return ParseAcceptLanguage(r.Header().Get("Accept-Language"))
}

// ParseAcceptLanguage 解析 HTTP Accept-Language 头。
func ParseAcceptLanguage(header string) []Locale {
	type candidate struct {
		locale Locale
		q      float64
		index  int
	}
	candidates := make([]candidate, 0)
	for index, item := range strings.Split(header, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		value, paramsText, _ := strings.Cut(item, ";")
		locale, ok := NewLocale(value)
		if !ok {
			continue
		}
		q := 1.0
		if paramsText != "" {
			for _, param := range strings.Split(paramsText, ";") {
				name, raw, ok := strings.Cut(strings.TrimSpace(param), "=")
				if !ok || !strings.EqualFold(name, "q") {
					continue
				}
				parsed, err := strconv.ParseFloat(raw, 64)
				if err == nil && parsed >= 0 && parsed <= 1 {
					q = parsed
				}
			}
		}
		if q == 0 {
			continue
		}
		candidates = append(candidates, candidate{locale: locale, q: q, index: index})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].q == candidates[j].q {
			return candidates[i].index < candidates[j].index
		}
		return candidates[i].q > candidates[j].q
	})
	result := make([]Locale, len(candidates))
	for i, item := range candidates {
		result[i] = item.locale
	}
	return result
}
