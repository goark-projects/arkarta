package servlet

// MappingType 表示 Servlet 路径映射命中的类型。
type MappingType uint8

const (
	// MappingUnknown 表示请求尚未命中任何 Servlet 映射。
	MappingUnknown MappingType = iota
	// MappingDefault 表示默认映射 "/"。
	MappingDefault
	// MappingExact 表示精确路径映射。
	MappingExact
	// MappingPrefix 表示路径前缀映射。
	MappingPrefix
	// MappingExtension 表示扩展名映射。
	MappingExtension
)

// RequestMapping 描述一次请求命中的 Servlet 映射信息。
type RequestMapping struct {
	pattern     string
	mappingType MappingType
	servletPath string
	pathInfo    string
}

func newRequestMapping(pattern string, mappingType MappingType, servletPath, pathInfo string) RequestMapping {
	return RequestMapping{
		pattern:     pattern,
		mappingType: mappingType,
		servletPath: servletPath,
		pathInfo:    pathInfo,
	}
}

// Pattern 返回声明的 Servlet 映射模式。
func (m RequestMapping) Pattern() string {
	return m.pattern
}

// Type 返回映射命中类型。
func (m RequestMapping) Type() MappingType {
	return m.mappingType
}

// ServletPath 返回映射对应的 Servlet 路径。
func (m RequestMapping) ServletPath() string {
	return m.servletPath
}

// PathInfo 返回 ServletPath 之后的剩余路径。
func (m RequestMapping) PathInfo() string {
	return m.pathInfo
}
