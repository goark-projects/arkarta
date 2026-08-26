package session

import (
	"crypto/rand"
	"encoding/base64"
)

const defaultIDBytes = 32

// IDGenerator 生成高熵会话 ID。
type IDGenerator func() (string, error)

// SecureID 生成 URL 安全的随机会话 ID。
func SecureID() (string, error) {
	buffer := make([]byte, defaultIDBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
