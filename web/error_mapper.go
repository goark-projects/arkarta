package web

import (
	"errors"
	"fmt"
	"net/http"

	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/validation"
)

// ErrorMapper 将处理错误映射为稳定的 Web 响应。
type ErrorMapper interface {
	MapError(ctx *Context, err error) Result
}

// ErrorMapperFunc 将普通函数适配为 ErrorMapper。
type ErrorMapperFunc func(ctx *Context, err error) Result

// MapError 执行底层错误映射函数。
func (f ErrorMapperFunc) MapError(ctx *Context, err error) Result {
	if f == nil {
		return DefaultErrorMapper{}.MapError(ctx, err)
	}
	return f(ctx, err)
}

// DefaultErrorMapper 是 Arkarta Web 默认错误响应映射器。
type DefaultErrorMapper struct{}

// MapError 将常见标准错误映射为 JSON 错误体。
func (DefaultErrorMapper) MapError(_ *Context, err error) Result {
	statusCode, body := errorResponse(err)
	return errorJSON(statusCode, body)
}

// ErrorResponse 表示统一错误响应体。
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody 表示统一错误载荷。
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ViolationDetail 表示校验失败明细。
type ViolationDetail struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func errorResponse(err error) (int, ErrorResponse) {
	if err == nil {
		return http.StatusInternalServerError, newErrorResponse("INTERNAL_ERROR", http.StatusText(http.StatusInternalServerError), nil)
	}

	var validationErr validation.ValidationError
	if errors.As(err, &validationErr) {
		return http.StatusUnprocessableEntity, newErrorResponse("VALIDATION_ERROR", "请求参数校验失败", violationDetails(validationErr.Result()))
	}
	if errors.Is(err, arkjson.ErrPayloadTooLarge) {
		return http.StatusRequestEntityTooLarge, newErrorResponse("PAYLOAD_TOO_LARGE", http.StatusText(http.StatusRequestEntityTooLarge), nil)
	}
	if errors.Is(err, ErrUnsupportedMediaType) {
		return http.StatusUnsupportedMediaType, newErrorResponse("UNSUPPORTED_MEDIA_TYPE", http.StatusText(http.StatusUnsupportedMediaType), nil)
	}
	var bindErr *BindError
	if errors.As(err, &bindErr) {
		return http.StatusBadRequest, newErrorResponse("BAD_REQUEST", "请求体格式非法", nil)
	}
	var statusErr servlet.StatusError
	if errors.As(err, &statusErr) {
		return statusErr.StatusCode(), newErrorResponse(statusCodeName(statusErr.StatusCode()), statusErr.PublicMessage(), nil)
	}
	return http.StatusInternalServerError, newErrorResponse("INTERNAL_ERROR", http.StatusText(http.StatusInternalServerError), nil)
}

func newErrorResponse(code, message string, details any) ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

func violationDetails(result validation.Result) []ViolationDetail {
	violations := result.Violations()
	details := make([]ViolationDetail, 0, len(violations))
	for _, violation := range violations {
		details = append(details, ViolationDetail{
			Path:    violation.Path(),
			Code:    violation.Code(),
			Message: violation.Message(),
		})
	}
	return details
}

func statusCodeName(statusCode int) string {
	return fmt.Sprintf("HTTP_%d", statusCode)
}
