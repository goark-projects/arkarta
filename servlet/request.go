package servlet

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// ErrNilHTTPRequest 表示构造请求时传入了空的标准库请求。
var ErrNilHTTPRequest = errors.New("arkarta/servlet: http request is nil")

// ErrNilRequestInput 表示容器未提供请求输入。
var ErrNilRequestInput = errors.New("arkarta/servlet: request input is nil")

// DispatchType 表示请求进入处理链的原因。
type DispatchType uint8

const (
	// DispatchRequest 表示客户端原始请求。
	DispatchRequest DispatchType = iota
	// DispatchForward 表示服务端转发。
	DispatchForward
	// DispatchInclude 表示服务端包含。
	DispatchInclude
	// DispatchError 表示错误分发。
	DispatchError
	// DispatchAsync 表示异步分发。
	DispatchAsync
)

// RequestOption 定制 Request 构造行为。
type RequestOption func(*Request)

// RequestInput 是容器构造 Servlet 请求时提供的传输层中立数据。
// Header、Body 和 Trailer 的所有权至少持续到请求处理完成。
type RequestInput struct {
	Context       context.Context
	Method        string
	Protocol      string
	Scheme        string
	Host          string
	RequestURI    string
	Path          string
	QueryString   string
	Header        Header
	Body          io.ReadCloser
	ContentLength int64
	RemoteAddr    string
	LocalAddr     string
	Trailer       Header
	TrailerReady  func() bool
}

// WithDispatchType 设置请求分发类型。
func WithDispatchType(dispatchType DispatchType) RequestOption {
	return func(req *Request) {
		req.dispatchType = dispatchType
	}
}

// WithRequestConnectionID 设置容器分配的连接 ID。
func WithRequestConnectionID(connectionID string) RequestOption {
	return func(req *Request) {
		req.connectionID = strings.TrimSpace(connectionID)
	}
}

// Request 表示容器传给应用的请求视图。
type Request struct {
	httpRequest   *http.Request
	ctx           context.Context
	method        string
	protocol      string
	scheme        string
	host          string
	header        Header
	body          io.ReadCloser
	contentLength int64
	remoteAddr    string
	localAddr     string
	trailer       Header
	trailerReady  func() bool
	dispatchType  DispatchType
	connectionID  string
	requestURI    string
	queryString   string
	contextPath   string
	path          string
	servletPath   string
	pathInfo      string
	mapping       RequestMapping

	parametersOnce sync.Once
	parameters     url.Values
	parametersErr  error

	mu                 sync.RWMutex
	attribute          map[string]any
	attributeListeners []RequestAttributeListener
}

// NewRequest 从标准库请求创建 Arkarta Servlet 请求。
func NewRequest(httpRequest *http.Request, options ...RequestOption) (*Request, error) {
	if httpRequest == nil {
		return nil, ErrNilHTTPRequest
	}
	input := requestInputFromHTTP(httpRequest)
	req, err := NewRequestFromInput(&input, options...)
	if err != nil {
		return nil, err
	}
	req.httpRequest = httpRequest
	return req, nil
}

// NewRequestFromInput 从传输层中立输入创建 Arkarta Servlet 请求。
func NewRequestFromInput(input *RequestInput, options ...RequestOption) (*Request, error) {
	if input == nil {
		return nil, ErrNilRequestInput
	}
	ctx := input.Context
	if ctx == nil {
		ctx = context.Background()
	}
	header := input.Header
	if header == nil {
		header = NewHeader()
	}
	body := input.Body
	if body == nil {
		body = io.NopCloser(strings.NewReader(""))
	}
	trailer := input.Trailer
	if trailer == nil {
		trailer = NewHeader()
	}
	req := &Request{
		ctx:           ctx,
		method:        input.Method,
		protocol:      input.Protocol,
		scheme:        input.Scheme,
		host:          input.Host,
		header:        header,
		body:          body,
		contentLength: input.ContentLength,
		remoteAddr:    input.RemoteAddr,
		localAddr:     input.LocalAddr,
		trailer:       trailer,
		trailerReady:  input.TrailerReady,
		dispatchType:  DispatchRequest,
		requestURI:    input.RequestURI,
		queryString:   input.QueryString,
		path:          input.Path,
		attribute:     make(map[string]any),
	}
	for _, option := range options {
		if option != nil {
			option(req)
		}
	}
	return req, nil
}

// Context 返回请求上下文。
func (r *Request) Context() context.Context {
	return r.ctx
}

// Method 返回 HTTP 方法。
func (r *Request) Method() string {
	return r.method
}

// Protocol 返回 HTTP 协议版本。
func (r *Request) Protocol() string {
	return r.protocol
}

// Scheme 返回请求协议。
func (r *Request) Scheme() string {
	return r.scheme
}

// Host 返回请求主机。
func (r *Request) Host() string {
	return r.host
}

// RequestURI 返回不含查询串的原始请求 URI 路径。
func (r *Request) RequestURI() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.requestURI
}

// RequestURL 返回不含查询串的完整请求 URL。
func (r *Request) RequestURL() string {
	return r.Scheme() + "://" + r.Host() + r.RequestURI()
}

// QueryString 返回当前分发视图下的原始查询串。
func (r *Request) QueryString() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.queryString
}

// ContextPath 返回 Web 应用上下文路径。
func (r *Request) ContextPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.contextPath
}

// Path 返回 URL 路径。
func (r *Request) Path() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.path
}

// Query 返回 URL 查询参数副本。
func (r *Request) Query() url.Values {
	values, err := url.ParseQuery(r.QueryString())
	if err != nil {
		return url.Values{}
	}
	return cloneURLValues(values)
}

// ServletPath 返回匹配到当前 Servlet 的路径前缀。
func (r *Request) ServletPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.servletPath
}

// PathInfo 返回除 ServletPath 外的剩余路径。
func (r *Request) PathInfo() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pathInfo
}

// Mapping 返回当前请求命中的 Servlet 映射信息。
func (r *Request) Mapping() RequestMapping {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mapping
}

// Header 返回请求头。
func (r *Request) Header() Header {
	return r.header
}

// Cookie 返回指定名称的 Cookie。
func (r *Request) Cookie(name string) (*http.Cookie, error) {
	return r.httpRequest.Cookie(name)
}

// Body 返回请求体读取器。
func (r *Request) Body() io.ReadCloser {
	return r.body
}

// ContentLength 返回请求体长度。
func (r *Request) ContentLength() int64 {
	return r.contentLength
}

// RemoteAddr 返回远端网络地址。
func (r *Request) RemoteAddr() string {
	return r.remoteAddr
}

// IsSecure 表示请求是否通过安全传输进入容器。
func (r *Request) IsSecure() bool {
	return r.Scheme() == "https"
}

// Attribute 返回请求属性。
func (r *Request) Attribute(key string) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.attribute[key]
	return value, ok
}

// SetAttribute 设置请求属性；传入 nil 会删除该属性。
func (r *Request) SetAttribute(key string, value any) {
	r.setAttribute(r.Context(), key, value)
}

// DispatchType 返回当前分发类型。
func (r *Request) DispatchType() DispatchType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.dispatchType
}

// HTTPRequest 返回底层标准库请求。
func (r *Request) HTTPRequest() *http.Request {
	return r.httpRequest
}

type dispatchSnapshot struct {
	path         string
	queryString  string
	dispatchType DispatchType
	servletPath  string
	pathInfo     string
	mapping      RequestMapping
}

func (r *Request) dispatchSnapshot() dispatchSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return dispatchSnapshot{
		path:         r.path,
		queryString:  r.queryString,
		dispatchType: r.dispatchType,
		servletPath:  r.servletPath,
		pathInfo:     r.pathInfo,
		mapping:      r.mapping,
	}
}

func (r *Request) applyDispatch(path, queryString string, dispatchType DispatchType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.path = path
	r.queryString = queryString
	r.dispatchType = dispatchType
}

func (r *Request) restoreDispatch(snapshot dispatchSnapshot) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.path = snapshot.path
	r.queryString = snapshot.queryString
	r.dispatchType = snapshot.dispatchType
	r.servletPath = snapshot.servletPath
	r.pathInfo = snapshot.pathInfo
	r.mapping = snapshot.mapping
}

func (r *Request) applyMapping(mapping RequestMapping) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servletPath = mapping.ServletPath()
	r.pathInfo = mapping.PathInfo()
	r.mapping = mapping
}

func requestPath(httpRequest *http.Request) string {
	if httpRequest.URL == nil {
		return ""
	}
	return httpRequest.URL.Path
}

func requestURI(httpRequest *http.Request) string {
	if httpRequest.URL == nil {
		return ""
	}
	if uri := httpRequest.URL.EscapedPath(); uri != "" {
		return uri
	}
	if httpRequest.URL.Path != "" {
		return httpRequest.URL.Path
	}
	return "/"
}

func requestQueryString(httpRequest *http.Request) string {
	if httpRequest.URL == nil {
		return ""
	}
	return httpRequest.URL.RawQuery
}

func requestInputFromHTTP(httpRequest *http.Request) RequestInput {
	scheme := "http"
	if httpRequest.URL != nil && httpRequest.URL.Scheme != "" {
		scheme = httpRequest.URL.Scheme
	} else if httpRequest.TLS != nil {
		scheme = "https"
	}
	localAddr := ""
	if addr, ok := httpRequest.Context().Value(http.LocalAddrContextKey).(interface{ String() string }); ok && addr != nil {
		localAddr = addr.String()
	}
	return RequestInput{
		Context:       httpRequest.Context(),
		Method:        httpRequest.Method,
		Protocol:      httpRequest.Proto,
		Scheme:        scheme,
		Host:          httpRequest.Host,
		RequestURI:    requestURI(httpRequest),
		Path:          requestPath(httpRequest),
		QueryString:   requestQueryString(httpRequest),
		Header:        mapHeader(httpRequest.Header),
		Body:          httpRequest.Body,
		ContentLength: httpRequest.ContentLength,
		RemoteAddr:    httpRequest.RemoteAddr,
		LocalAddr:     localAddr,
		Trailer:       mapHeader(httpRequest.Trailer),
		TrailerReady: func() bool {
			for _, values := range httpRequest.Trailer {
				if values == nil {
					return false
				}
			}
			return true
		},
	}
}

// WithRequestContextPath 设置请求所属 Web 应用的上下文路径。
func WithRequestContextPath(contextPath string) RequestOption {
	return func(req *Request) {
		req.contextPath = normalizeRequestContextPath(contextPath)
		req.path = stripRequestContextPath(req.path, req.contextPath)
		req.servletPath = ""
		req.pathInfo = ""
		req.mapping = RequestMapping{}
	}
}

func normalizeRequestContextPath(contextPath string) string {
	if contextPath == "" || contextPath == "/" {
		return ""
	}
	if !strings.HasPrefix(contextPath, "/") {
		contextPath = "/" + contextPath
	}
	return strings.TrimRight(contextPath, "/")
}

func stripRequestContextPath(path, contextPath string) string {
	if path == "" {
		return ""
	}
	if contextPath == "" {
		return path
	}
	if path == contextPath {
		return "/"
	}
	if strings.HasPrefix(path, contextPath+"/") {
		return strings.TrimPrefix(path, contextPath)
	}
	return path
}
