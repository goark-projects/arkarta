package servlet

const (
	// AttributeForwardRequestURI 保存 forward 前的原始请求路径。
	AttributeForwardRequestURI = "arkarta.servlet.forward.request_uri"
	// AttributeIncludeRequestURI 保存 include 前的原始请求路径。
	AttributeIncludeRequestURI = "arkarta.servlet.include.request_uri"
	// AttributeErrorStatusCode 保存错误分发的 HTTP 状态码。
	AttributeErrorStatusCode = "arkarta.servlet.error.status_code"
	// AttributeErrorException 保存错误分发的错误对象。
	AttributeErrorException = "arkarta.servlet.error.exception"
	// AttributeErrorRequestURI 保存错误发生时的请求路径。
	AttributeErrorRequestURI = "arkarta.servlet.error.request_uri"
)
