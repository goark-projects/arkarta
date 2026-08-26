package registration

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
		dst[i].initParam = cloneStringMap(dst[i].initParam)
		dst[i].mappings = cloneStrings(dst[i].mappings)
	}
	return dst
}

// Filters 返回 Filter 元数据副本。
func (s Snapshot) Filters() []FilterDescriptor {
	if len(s.filters) == 0 {
		return nil
	}
	dst := make([]FilterDescriptor, len(s.filters))
	copy(dst, s.filters)
	for i := range dst {
		dst[i].initParam = cloneStringMap(dst[i].initParam)
		dst[i].urlPatternMappings = cloneURLPatternMappings(dst[i].urlPatternMappings)
		dst[i].servletNameMappings = cloneServletNameMappings(dst[i].servletNameMappings)
	}
	return dst
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
