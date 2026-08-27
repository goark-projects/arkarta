package validation

import "strings"

// Violation 表示一次字段或对象校验失败。
type Violation struct {
	path    string
	code    string
	message string
	value   any
}

// NewViolation 创建校验失败项。
func NewViolation(path, code, message string, value any) Violation {
	return Violation{
		path:    strings.TrimSpace(path),
		code:    strings.TrimSpace(code),
		message: strings.TrimSpace(message),
		value:   value,
	}
}

// Path 返回字段路径。
func (v Violation) Path() string {
	return v.path
}

// Code 返回机器可读错误码。
func (v Violation) Code() string {
	return v.code
}

// Message 返回安全错误信息。
func (v Violation) Message() string {
	return v.message
}

// Value 返回导致校验失败的值。
func (v Violation) Value() any {
	return v.value
}
