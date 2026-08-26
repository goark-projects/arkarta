package multipart

const defaultMaxMemory = 32 << 20

// Option 定制 Multipart 解析器。
type Option func(*Parser)

// WithMaxMemory 设置内存解析阈值。
func WithMaxMemory(maxMemory int64) Option {
	return func(parser *Parser) {
		if maxMemory > 0 {
			parser.maxMemory = maxMemory
		}
	}
}

// WithMaxBodySize 设置最大请求体大小。
func WithMaxBodySize(maxBodySize int64) Option {
	return func(parser *Parser) {
		parser.maxBodySize = maxBodySize
	}
}

// WithParserMaxFileSize 设置单个文件大小上限。
func WithParserMaxFileSize(maxFileSize int64) Option {
	return func(parser *Parser) {
		parser.maxFileSize = maxFileSize
	}
}

// WithConfig 使用 multipart 配置创建解析器选项。
func WithConfig(config Config) Option {
	return func(parser *Parser) {
		for _, option := range config.ParserOptions() {
			option(parser)
		}
	}
}
