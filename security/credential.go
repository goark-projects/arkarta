package security

import "strings"

const (
	// CredentialPassword 表示用户名密码凭证。
	CredentialPassword = "password"
	// CredentialBearer 表示 Bearer Token 凭证。
	CredentialBearer = "bearer"
)

// Credential 表示可交给认证管理器处理的凭证。
type Credential interface {
	Kind() string
}

// PasswordCredential 表示用户名密码凭证。
type PasswordCredential struct {
	username string
	password string
}

// NewPasswordCredential 创建用户名密码凭证。
func NewPasswordCredential(username, password string) PasswordCredential {
	return PasswordCredential{
		username: strings.TrimSpace(username),
		password: password,
	}
}

// Kind 返回凭证类型。
func (c PasswordCredential) Kind() string {
	return CredentialPassword
}

// Username 返回用户名。
func (c PasswordCredential) Username() string {
	return c.username
}

// Password 返回密码明文。
func (c PasswordCredential) Password() string {
	return c.password
}

// BearerCredential 表示 Bearer Token 凭证。
type BearerCredential struct {
	token string
}

// NewBearerCredential 创建 Bearer Token 凭证。
func NewBearerCredential(token string) BearerCredential {
	return BearerCredential{token: strings.TrimSpace(token)}
}

// Kind 返回凭证类型。
func (c BearerCredential) Kind() string {
	return CredentialBearer
}

// Token 返回 Bearer Token。
func (c BearerCredential) Token() string {
	return c.token
}
