package validation

import (
	"fmt"
	"strings"
)

// Result 表示一次校验结果。
type Result struct {
	violations []Violation
}

// NewResult 创建校验结果。
func NewResult(violations ...Violation) Result {
	return Result{violations: cloneViolations(violations)}
}

// Valid 表示校验是否通过。
func (r Result) Valid() bool {
	return len(r.violations) == 0
}

// Violations 返回校验失败项副本。
func (r Result) Violations() []Violation {
	return cloneViolations(r.violations)
}

// Error 转换为标准 error；校验通过时返回 nil。
func (r Result) Error() error {
	if r.Valid() {
		return nil
	}
	return ValidationError{result: r}
}

// ValidationError 表示携带 Result 的校验错误。
type ValidationError struct {
	result Result
}

// Error 返回简洁错误摘要。
func (e ValidationError) Error() string {
	violations := e.result.Violations()
	if len(violations) == 0 {
		return "validation failed"
	}
	first := violations[0]
	if first.Path() == "" {
		return fmt.Sprintf("validation failed: %s", first.Message())
	}
	return fmt.Sprintf("validation failed: %s %s", first.Path(), first.Message())
}

// Result 返回完整校验结果。
func (e ValidationError) Result() Result {
	return e.result
}

func appendViolation(violations []Violation, path, code, message string, value any) []Violation {
	if message == "" {
		message = strings.TrimSpace(code)
	}
	return append(violations, NewViolation(path, code, message, value))
}

func cloneViolations(src []Violation) []Violation {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Violation, len(src))
	copy(dst, src)
	return dst
}
