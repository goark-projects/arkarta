package frame

// MaskKey 表示 WebSocket 客户端帧的 4 字节掩码。
type MaskKey [4]byte

// MaskPayload 返回按 MaskKey 异或后的载荷副本。
func MaskPayload(payload []byte, key MaskKey) []byte {
	dst := cloneBytes(payload)
	ApplyMask(dst, key)
	return dst
}

// ApplyMask 原地执行 WebSocket Mask 异或。
func ApplyMask(payload []byte, key MaskKey) {
	for i := range payload {
		payload[i] ^= key[i&3]
	}
}
