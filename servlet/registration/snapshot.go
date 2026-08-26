package registration

import "sort"

// Snapshot 是注册表的不可变视图。
type Snapshot struct {
	frozen    bool
	servlets  []ServletDescriptor
	filters   []FilterDescriptor
	listeners []ListenerDescriptor
}

// Frozen 表示快照创建时注册表是否已冻结。
func (s Snapshot) Frozen() bool {
	return s.frozen
}

// Servlets 返回 Servlet 元数据副本。
func (s Snapshot) Servlets() []ServletDescriptor {
	if len(s.servlets) == 0 {
		return nil
	}
	dst := make([]ServletDescriptor, len(s.servlets))
	copy(dst, s.servlets)
	for i := range dst {
		dst[i] = dst[i].clone()
	}
	return dst
}

// Servlet 按名称查询 Servlet 快照。
func (s Snapshot) Servlet(name string) (ServletDescriptor, bool) {
	for _, descriptor := range s.servlets {
		if descriptor.name == name {
			return descriptor.clone(), true
		}
	}
	return ServletDescriptor{}, false
}

// ServletNames 返回已注册 Servlet 名称集合。
func (s Snapshot) ServletNames() []string {
	result := make([]string, 0, len(s.servlets))
	for _, descriptor := range s.servlets {
		result = append(result, descriptor.name)
	}
	sort.Strings(result)
	return result
}

// Filters 返回 Filter 元数据副本。
func (s Snapshot) Filters() []FilterDescriptor {
	if len(s.filters) == 0 {
		return nil
	}
	dst := make([]FilterDescriptor, len(s.filters))
	copy(dst, s.filters)
	for i := range dst {
		dst[i] = dst[i].clone()
	}
	return dst
}

// Filter 按名称查询 Filter 快照。
func (s Snapshot) Filter(name string) (FilterDescriptor, bool) {
	for _, descriptor := range s.filters {
		if descriptor.name == name {
			return descriptor.clone(), true
		}
	}
	return FilterDescriptor{}, false
}

// FilterNames 返回已注册 Filter 名称集合。
func (s Snapshot) FilterNames() []string {
	result := make([]string, 0, len(s.filters))
	for _, descriptor := range s.filters {
		result = append(result, descriptor.name)
	}
	sort.Strings(result)
	return result
}

// Listeners 返回 Listener 元数据副本。
func (s Snapshot) Listeners() []ListenerDescriptor {
	if len(s.listeners) == 0 {
		return nil
	}
	dst := make([]ListenerDescriptor, len(s.listeners))
	copy(dst, s.listeners)
	return dst
}

// ListenersByKind 返回指定类型的 Listener 快照。
func (s Snapshot) ListenersByKind(kind ListenerKind) []ListenerDescriptor {
	result := make([]ListenerDescriptor, 0)
	for _, descriptor := range s.listeners {
		if descriptor.kind == kind {
			result = append(result, descriptor)
		}
	}
	return result
}

func (r *Registry) snapshotLocked() Snapshot {
	snapshot := Snapshot{
		frozen:    r.frozen,
		servlets:  make([]ServletDescriptor, 0, len(r.servletOrder)),
		filters:   make([]FilterDescriptor, 0, len(r.filterOrder)),
		listeners: make([]ListenerDescriptor, 0, len(r.listeners)),
	}
	for _, name := range r.servletOrder {
		snapshot.servlets = append(snapshot.servlets, r.servlets[name].descriptorLocked())
	}
	for _, name := range r.filterOrder {
		snapshot.filters = append(snapshot.filters, r.filters[name].descriptorLocked())
	}
	for _, listener := range r.listeners {
		snapshot.listeners = append(snapshot.listeners, listener.descriptorLocked())
	}
	return snapshot
}
