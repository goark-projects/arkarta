package websocket

import (
	"bytes"
	"compress/flate"
	"io"
)

// ExtensionPerMessageDeflate 是 RFC 7692 定义的消息压缩扩展名称。
const ExtensionPerMessageDeflate = "permessage-deflate"

var perMessageDeflateTail = []byte{0x00, 0x00, 0xff, 0xff}
var perMessageDeflateFinalBlock = []byte{0x01, 0x00, 0x00, 0xff, 0xff}

// PerMessageDeflateOption 定制 permessage-deflate 协商策略。
type PerMessageDeflateOption func(*PerMessageDeflate)

// PerMessageDeflate 表示 permessage-deflate 扩展协商与压缩辅助。
type PerMessageDeflate struct {
	serverNoContextTakeover bool
	clientNoContextTakeover bool
	level                   int
}

// NewPerMessageDeflate 创建 permessage-deflate 扩展协商器。
func NewPerMessageDeflate(options ...PerMessageDeflateOption) *PerMessageDeflate {
	extension := &PerMessageDeflate{level: flate.BestSpeed}
	for _, option := range options {
		if option != nil {
			option(extension)
		}
	}
	return extension
}

// WithServerNoContextTakeover 要求服务端消息之间不复用压缩上下文。
func WithServerNoContextTakeover(enabled bool) PerMessageDeflateOption {
	return func(extension *PerMessageDeflate) {
		extension.serverNoContextTakeover = enabled
	}
}

// WithClientNoContextTakeover 要求客户端消息之间不复用压缩上下文。
func WithClientNoContextTakeover(enabled bool) PerMessageDeflateOption {
	return func(extension *PerMessageDeflate) {
		extension.clientNoContextTakeover = enabled
	}
}

// WithCompressionLevel 设置 DEFLATE 压缩级别。
func WithCompressionLevel(level int) PerMessageDeflateOption {
	return func(extension *PerMessageDeflate) {
		if level >= flate.HuffmanOnly && level <= flate.BestCompression {
			extension.level = level
		}
	}
}

// Negotiate 从客户端扩展列表中协商 permessage-deflate。
func (p *PerMessageDeflate) Negotiate(offers []Extension) (Extension, bool) {
	for _, offer := range offers {
		if offer.Name() != ExtensionPerMessageDeflate {
			continue
		}
		params := make(map[string]string)
		if p != nil && p.serverNoContextTakeover {
			params["server_no_context_takeover"] = ""
		}
		if p != nil && p.clientNoContextTakeover {
			params["client_no_context_takeover"] = ""
		}
		extension, ok := NewExtension(ExtensionPerMessageDeflate, params)
		return extension, ok
	}
	return Extension{}, false
}

// CompressMessage 使用 permessage-deflate 线格式压缩单条消息。
func (p *PerMessageDeflate) CompressMessage(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	level := flate.BestSpeed
	if p != nil {
		level = p.level
	}
	writer, err := flate.NewWriter(&buffer, level)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		_ = writer.Close()
		return nil, err
	}
	compressed := buffer.Bytes()
	if bytes.HasSuffix(compressed, perMessageDeflateTail) {
		compressed = compressed[:len(compressed)-len(perMessageDeflateTail)]
	}
	result := append([]byte(nil), compressed...)
	_ = writer.Close()
	return result, nil
}

// DecompressMessage 解压 permessage-deflate 单条消息；maxSize 小于 0 表示不限制。
func (p *PerMessageDeflate) DecompressMessage(data []byte, maxSize int64) ([]byte, error) {
	payload := make([]byte, 0, len(data)+len(perMessageDeflateTail)+len(perMessageDeflateFinalBlock))
	payload = append(payload, data...)
	payload = append(payload, perMessageDeflateTail...)
	payload = append(payload, perMessageDeflateFinalBlock...)
	reader := flate.NewReader(bytes.NewReader(payload))
	defer reader.Close()

	var buffer bytes.Buffer
	writer := io.Writer(&buffer)
	if maxSize >= 0 {
		writer = &limitWriter{writer: writer, remaining: maxSize}
	}
	if _, err := io.Copy(writer, reader); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type limitWriter struct {
	writer    io.Writer
	remaining int64
}

func (w *limitWriter) Write(data []byte) (int, error) {
	if int64(len(data)) > w.remaining {
		if w.remaining > 0 {
			limit := int(w.remaining)
			n, err := w.writer.Write(data[:limit])
			w.remaining = 0
			if err != nil {
				return n, err
			}
			return n, ErrMessageTooLarge
		}
		return 0, ErrMessageTooLarge
	}
	n, err := w.writer.Write(data)
	w.remaining -= int64(n)
	return n, err
}
