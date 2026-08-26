package multipart

// Config 描述 multipart/form-data 解析约束。
type Config struct {
	location          string
	maxFileSize       int64
	maxRequestSize    int64
	fileSizeThreshold int64
}

// ConfigOption 定制 multipart 配置。
type ConfigOption func(*Config)

// NewConfig 创建 multipart 配置。
func NewConfig(options ...ConfigOption) Config {
	config := Config{
		maxFileSize:       -1,
		maxRequestSize:    -1,
		fileSizeThreshold: defaultMaxMemory,
	}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	return config
}

// WithLocation 设置容器用于临时文件的目录。
func WithLocation(location string) ConfigOption {
	return func(config *Config) {
		config.location = location
	}
}

// WithMaxFileSize 设置单个上传文件的最大字节数。
func WithMaxFileSize(maxFileSize int64) ConfigOption {
	return func(config *Config) {
		config.maxFileSize = maxFileSize
	}
}

// WithMaxRequestSize 设置 multipart 请求体最大字节数。
func WithMaxRequestSize(maxRequestSize int64) ConfigOption {
	return func(config *Config) {
		config.maxRequestSize = maxRequestSize
	}
}

// WithFileSizeThreshold 设置落盘前的内存阈值。
func WithFileSizeThreshold(threshold int64) ConfigOption {
	return func(config *Config) {
		if threshold > 0 {
			config.fileSizeThreshold = threshold
		}
	}
}

// Location 返回临时文件目录。
func (c Config) Location() string {
	return c.location
}

// MaxFileSize 返回单个文件大小上限；负数表示不限。
func (c Config) MaxFileSize() int64 {
	return c.maxFileSize
}

// MaxRequestSize 返回请求体大小上限；负数表示不限。
func (c Config) MaxRequestSize() int64 {
	return c.maxRequestSize
}

// FileSizeThreshold 返回落盘前的内存阈值。
func (c Config) FileSizeThreshold() int64 {
	if c.fileSizeThreshold <= 0 {
		return defaultMaxMemory
	}
	return c.fileSizeThreshold
}

// ParserOptions 将配置转换为标准解析器选项。
func (c Config) ParserOptions() []Option {
	return []Option{
		WithMaxMemory(c.FileSizeThreshold()),
		WithMaxBodySize(c.MaxRequestSize()),
		WithParserMaxFileSize(c.MaxFileSize()),
		WithParserLocation(c.Location()),
	}
}
