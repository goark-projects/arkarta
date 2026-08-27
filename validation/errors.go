package validation

import "errors"

// ErrNilValidator 表示校验入口为空。
var ErrNilValidator = errors.New("arkarta/validation: validator is nil")

// ErrNilValue 表示校验目标为空。
var ErrNilValue = errors.New("arkarta/validation: value is nil")

// ErrInvalidRule 表示校验规则声明非法。
var ErrInvalidRule = errors.New("arkarta/validation: invalid rule")

// ErrUnsupportedValue 表示校验目标类型不受支持。
var ErrUnsupportedValue = errors.New("arkarta/validation: unsupported value")
