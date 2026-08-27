package security

// Principal 表示认证主体。
type Principal interface {
	Name() string
}

// PrincipalFunc 将普通函数适配为 Principal。
type PrincipalFunc func() string

// Name 返回主体名称。
func (f PrincipalFunc) Name() string {
	if f == nil {
		return ""
	}
	return f()
}

// Authority 表示主体拥有的一项权限。
type Authority string
