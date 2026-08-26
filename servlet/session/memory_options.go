package session

import "time"

// MemoryManagerOption 定制内存会话管理器。
type MemoryManagerOption func(*MemoryManager)

// WithIDGenerator 设置会话 ID 生成器。
func WithIDGenerator(generator IDGenerator) MemoryManagerOption {
	return func(manager *MemoryManager) {
		if generator != nil {
			manager.idGenerator = generator
		}
	}
}

// WithClock 设置时间源，测试中用于控制过期行为。
func WithClock(clock func() time.Time) MemoryManagerOption {
	return func(manager *MemoryManager) {
		if clock != nil {
			manager.clock = clock
		}
	}
}

// WithMaxInactiveInterval 设置默认空闲超时。
func WithMaxInactiveInterval(interval time.Duration) MemoryManagerOption {
	return func(manager *MemoryManager) {
		manager.maxInactiveInterval = interval
	}
}
