package nativeio

import "sort"

// Capability 表示容器暴露的原生 I/O 能力。
type Capability string

const (
	// CapabilitySendfile 表示容器支持文件到连接的内核级发送路径。
	CapabilitySendfile Capability = "sendfile"
	// CapabilitySplice 表示容器支持管道或描述符间的零拷贝搬运。
	CapabilitySplice Capability = "splice"
	// CapabilityIOUring 表示容器支持 Linux io_uring。
	CapabilityIOUring Capability = "io_uring"
	// CapabilityEpoll 表示容器支持 Linux epoll。
	CapabilityEpoll Capability = "epoll"
	// CapabilityKqueue 表示容器支持 BSD/macOS kqueue。
	CapabilityKqueue Capability = "kqueue"
	// CapabilityBackpressure 表示容器能暴露连接背压信号。
	CapabilityBackpressure Capability = "backpressure"
)

// Capabilities 保存 Native I/O 能力集合。
type Capabilities struct {
	values map[Capability]struct{}
}

// NewCapabilities 创建能力集合。
func NewCapabilities(values ...Capability) Capabilities {
	result := Capabilities{values: make(map[Capability]struct{}, len(values))}
	for _, value := range values {
		if value != "" {
			result.values[value] = struct{}{}
		}
	}
	return result
}

// Has 判断能力集合是否包含指定能力。
func (c Capabilities) Has(value Capability) bool {
	if value == "" {
		return false
	}
	_, ok := c.values[value]
	return ok
}

// Values 返回稳定排序的能力列表。
func (c Capabilities) Values() []Capability {
	result := make([]Capability, 0, len(c.values))
	for value := range c.values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}
