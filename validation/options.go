package validation

// Option 定制默认结构体验证器。
type Option func(*DefaultValidator)

// WithTagName 设置结构体标签名称。
func WithTagName(name string) Option {
	return func(validator *DefaultValidator) {
		if name != "" {
			validator.tagName = name
		}
	}
}

// WithConstraint 注册或覆盖字段级约束。
func WithConstraint(constraint Constraint) Option {
	return func(validator *DefaultValidator) {
		validator.register(constraint)
	}
}
