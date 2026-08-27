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

// WithMessageResolver 设置校验失败消息解析器。
func WithMessageResolver(resolver MessageResolver) Option {
	return func(validator *DefaultValidator) {
		if resolver != nil {
			validator.messageResolver = resolver
		}
	}
}

// WithObjectConstraint 注册默认分组的对象级约束。
func WithObjectConstraint(sample any, constraint ObjectConstraint) Option {
	return WithObjectConstraintForGroups(sample, constraint, DefaultGroup)
}

// WithObjectConstraintForGroups 注册指定分组的对象级约束。
func WithObjectConstraintForGroups(sample any, constraint ObjectConstraint, groups ...string) Option {
	return func(validator *DefaultValidator) {
		validator.registerObjectConstraint(sample, constraint, groups)
	}
}
