package frame

import (
	"encoding/binary"
	"unicode/utf8"
)

// ClosePayload 创建关闭帧载荷。
func ClosePayload(code uint16, reason string) ([]byte, error) {
	if code == 0 && reason == "" {
		return nil, nil
	}
	if !validCloseCode(code) || !utf8.ValidString(reason) {
		return nil, ErrInvalidClosePayload
	}
	payload := make([]byte, 2+len(reason))
	binary.BigEndian.PutUint16(payload[:2], code)
	copy(payload[2:], reason)
	if len(payload) > 125 {
		return nil, ErrInvalidClosePayload
	}
	return payload, nil
}

// ParseClosePayload 解析关闭帧状态码和原因。
func ParseClosePayload(payload []byte) (uint16, string, error) {
	if len(payload) == 0 {
		return 0, "", nil
	}
	if len(payload) == 1 {
		return 0, "", ErrInvalidClosePayload
	}
	code := binary.BigEndian.Uint16(payload[:2])
	reason := string(payload[2:])
	if !validCloseCode(code) || !utf8.ValidString(reason) {
		return 0, "", ErrInvalidClosePayload
	}
	return code, reason, nil
}

func validCloseCode(code uint16) bool {
	if code < 1000 || code >= 5000 {
		return false
	}
	switch code {
	case 1004, 1005, 1006, 1015:
		return false
	}
	return true
}
