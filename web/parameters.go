package web

import (
	"fmt"
	"strconv"
)

// FormValue 返回 Servlet 请求参数中的第一个值。
func (c *Context) FormValue(name string) (string, bool, error) {
	if c == nil || c.request == nil {
		return "", false, ErrNilContext
	}
	return c.request.Parameter(name)
}

// FormValues 返回 Servlet 请求参数中的全部值副本。
func (c *Context) FormValues(name string) ([]string, bool, error) {
	if c == nil || c.request == nil {
		return nil, false, ErrNilContext
	}
	return c.request.ParameterValues(name)
}

// QueryInt 将查询参数转换为 int。
func (c *Context) QueryInt(name string) (int, bool, error) {
	if c == nil || c.request == nil {
		return 0, false, ErrNilContext
	}
	return parseIntParameter(name, c.request.Query().Get(name))
}

// QueryInt64 将查询参数转换为 int64。
func (c *Context) QueryInt64(name string) (int64, bool, error) {
	if c == nil || c.request == nil {
		return 0, false, ErrNilContext
	}
	return parseInt64Parameter(name, c.request.Query().Get(name))
}

// QueryBool 将查询参数转换为 bool。
func (c *Context) QueryBool(name string) (bool, bool, error) {
	if c == nil || c.request == nil {
		return false, false, ErrNilContext
	}
	return parseBoolParameter(name, c.request.Query().Get(name))
}

// PathInt 将路径变量转换为 int。
func (c *Context) PathInt(name string) (int, bool, error) {
	value, ok := c.Param(name)
	if !ok {
		return 0, false, nil
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, true, newParameterError(name, value, "int", err)
	}
	return result, true, nil
}

// PathInt64 将路径变量转换为 int64。
func (c *Context) PathInt64(name string) (int64, bool, error) {
	value, ok := c.Param(name)
	if !ok {
		return 0, false, nil
	}
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, true, newParameterError(name, value, "int64", err)
	}
	return result, true, nil
}

// PathBool 将路径变量转换为 bool。
func (c *Context) PathBool(name string) (bool, bool, error) {
	value, ok := c.Param(name)
	if !ok {
		return false, false, nil
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false, true, newParameterError(name, value, "bool", err)
	}
	return result, true, nil
}

func parseIntParameter(name, value string) (int, bool, error) {
	if value == "" {
		return 0, false, nil
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, true, newParameterError(name, value, "int", err)
	}
	return result, true, nil
}

func parseInt64Parameter(name, value string) (int64, bool, error) {
	if value == "" {
		return 0, false, nil
	}
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, true, newParameterError(name, value, "int64", err)
	}
	return result, true, nil
}

func parseBoolParameter(name, value string) (bool, bool, error) {
	if value == "" {
		return false, false, nil
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false, true, newParameterError(name, value, "bool", err)
	}
	return result, true, nil
}

func newParameterError(name, value, targetType string, cause error) *ParameterError {
	return &ParameterError{
		Name:  name,
		Value: value,
		Type:  targetType,
		Cause: fmt.Errorf("%w: %v", ErrInvalidParameter, cause),
	}
}
