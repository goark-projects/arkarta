package servlet

import "strings"

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
