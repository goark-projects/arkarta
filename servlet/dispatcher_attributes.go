package servlet

const (
	// AttributeServletName 保存当前请求命中的 Servlet 名称。
	AttributeServletName = "arkarta.servlet.servlet_name"
	// AttributeForwardRequestURI 保存 forward 前的原始请求路径。
	AttributeForwardRequestURI = "arkarta.servlet.forward.request_uri"
	// AttributeForwardContextPath 保存 forward 前的上下文路径。
	AttributeForwardContextPath = "arkarta.servlet.forward.context_path"
	// AttributeForwardServletPath 保存 forward 前的 Servlet 路径。
	AttributeForwardServletPath = "arkarta.servlet.forward.servlet_path"
	// AttributeForwardPathInfo 保存 forward 前的 PathInfo。
	AttributeForwardPathInfo = "arkarta.servlet.forward.path_info"
	// AttributeForwardQueryString 保存 forward 前的查询串。
	AttributeForwardQueryString = "arkarta.servlet.forward.query_string"
	// AttributeForwardMapping 保存 forward 前的映射信息。
	AttributeForwardMapping = "arkarta.servlet.forward.mapping"
	// AttributeIncludeRequestURI 保存 include 前的原始请求路径。
	AttributeIncludeRequestURI = "arkarta.servlet.include.request_uri"
	// AttributeIncludeContextPath 保存 include 前的上下文路径。
	AttributeIncludeContextPath = "arkarta.servlet.include.context_path"
	// AttributeIncludeServletPath 保存 include 前的 Servlet 路径。
	AttributeIncludeServletPath = "arkarta.servlet.include.servlet_path"
	// AttributeIncludePathInfo 保存 include 前的 PathInfo。
	AttributeIncludePathInfo = "arkarta.servlet.include.path_info"
	// AttributeIncludeQueryString 保存 include 前的查询串。
	AttributeIncludeQueryString = "arkarta.servlet.include.query_string"
	// AttributeIncludeMapping 保存 include 前的映射信息。
	AttributeIncludeMapping = "arkarta.servlet.include.mapping"
	// AttributeErrorStatusCode 保存错误分发的 HTTP 状态码。
	AttributeErrorStatusCode = "arkarta.servlet.error.status_code"
	// AttributeErrorException 保存错误分发的错误对象。
	AttributeErrorException = "arkarta.servlet.error.exception"
	// AttributeErrorExceptionType 保存错误分发的错误类型名称。
	AttributeErrorExceptionType = "arkarta.servlet.error.exception_type"
	// AttributeErrorMessage 保存错误分发的公开错误消息。
	AttributeErrorMessage = "arkarta.servlet.error.message"
	// AttributeErrorRequestURI 保存错误发生时的请求路径。
	AttributeErrorRequestURI = "arkarta.servlet.error.request_uri"
	// AttributeErrorQueryString 保存错误发生时的查询串。
	AttributeErrorQueryString = "arkarta.servlet.error.query_string"
	// AttributeErrorServletName 保存错误发生时的 Servlet 名称。
	AttributeErrorServletName = "arkarta.servlet.error.servlet_name"
)
