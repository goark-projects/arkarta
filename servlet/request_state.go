package servlet

import (
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// RunWithDispatchType 在指定分发类型下执行函数并恢复请求状态。
func (r *Request) RunWithDispatchType(dispatchType DispatchType, fn func() error) error {
	snapshot := r.dispatchSnapshot()
	r.applyDispatch(snapshot.path, snapshot.queryString, dispatchType)
	defer r.restoreDispatch(snapshot)
	if fn == nil {
		return nil
	}
	return fn()
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

// SetMethod 替换过滤器链后续处理器看到的请求方法。
func (r *Request) SetMethod(method string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.method = method
	if r.httpRequest != nil {
		r.httpRequest.Method = method
	}
}

// SetScheme 替换过滤器链后续处理器看到的请求协议。
func (r *Request) SetScheme(scheme string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scheme = scheme
	if r.httpRequest != nil && r.httpRequest.URL != nil {
		r.httpRequest.URL.Scheme = scheme
	}
}

// SetHost 替换过滤器链后续处理器看到的请求主机。
func (r *Request) SetHost(host string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.host = host
	if r.httpRequest != nil {
		r.httpRequest.Host = host
		if r.httpRequest.URL != nil {
			r.httpRequest.URL.Host = host
		}
	}
}

// SetRemoteAddr 替换过滤器链后续处理器看到的远端地址。
func (r *Request) SetRemoteAddr(address string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.remoteAddr = address
	if r.httpRequest != nil {
		r.httpRequest.RemoteAddr = address
	}
}

// SetBody 替换请求体及其长度；空请求体会被规范化为空读取器。
func (r *Request) SetBody(body io.ReadCloser, contentLength int64) {
	if r == nil {
		return
	}
	if body == nil {
		body = io.NopCloser(strings.NewReader(""))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.body = body
	r.contentLength = contentLength
	if r.httpRequest != nil {
		r.httpRequest.Body = body
		r.httpRequest.ContentLength = contentLength
	}
}

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
	return r.localAddr
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

// ErrFormBodyTooLarge 表示 URL 编码表单体超过标准解析上限。
var ErrFormBodyTooLarge = errors.New("arkarta/servlet: form body too large")

// ParseParameters 解析并缓存请求参数。
func (r *Request) ParseParameters() error {
	r.parametersOnce.Do(func() {
		r.parameters, r.parametersErr = r.readParameters()
	})
	return r.parametersErr
}

// Parameters 返回查询串和表单体合并后的请求参数副本。
func (r *Request) Parameters() (url.Values, error) {
	if err := r.ParseParameters(); err != nil {
		return nil, err
	}
	return cloneURLValues(r.parameters), nil
}

// Parameter 返回指定请求参数的第一个值。
func (r *Request) Parameter(name string) (string, bool, error) {
	values, err := r.Parameters()
	if err != nil {
		return "", false, err
	}
	list, ok := values[name]
	if !ok || len(list) == 0 {
		return "", false, nil
	}
	return list[0], true, nil
}

// ParameterValues 返回指定请求参数的全部值副本。
func (r *Request) ParameterValues(name string) ([]string, bool, error) {
	values, err := r.Parameters()
	if err != nil {
		return nil, false, err
	}
	list, ok := values[name]
	if !ok {
		return nil, false, nil
	}
	return append([]string(nil), list...), true, nil
}

// ParameterNames 返回请求参数名称的稳定排序副本。
func (r *Request) ParameterNames() ([]string, error) {
	values, err := r.Parameters()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func (r *Request) readParameters() (url.Values, error) {
	values, err := url.ParseQuery(r.QueryString())
	if err != nil {
		return nil, err
	}
	if shouldParseFormParameters(r) {
		body, err := readFormBody(r.Body(), r.maxFormBodySize)
		if err != nil {
			return nil, err
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		for name, list := range form {
			values[name] = append(values[name], list...)
		}
	}
	return values, nil
}

func readFormBody(body io.Reader, limit int64) ([]byte, error) {
	if limit < 0 {
		return io.ReadAll(body)
	}
	content, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(content)) > limit {
		return nil, ErrFormBodyTooLarge
	}
	return content, nil
}

func shouldParseFormParameters(request *Request) bool {
	if request == nil || request.Body() == nil {
		return false
	}
	switch request.Method() {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	default:
		return false
	}
	mediaType, _, err := mime.ParseMediaType(request.Header().Get("Content-Type"))
	if err != nil {
		return false
	}
	return strings.EqualFold(mediaType, "application/x-www-form-urlencoded")
}

func cloneURLValues(src url.Values) url.Values {
	if len(src) == 0 {
		return url.Values{}
	}
	dst := make(url.Values, len(src))
	for key, values := range src {
		dst[key] = append([]string(nil), values...)
	}
	return dst
}
