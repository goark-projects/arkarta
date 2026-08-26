package registration

import "goark.dev/arkarta/servlet"

// FilterRegistration 保存单个 Filter 的动态注册信息。
type FilterRegistration struct {
	commonRegistration

	filter              servlet.Filter
	urlPatternMappings  []URLPatternMapping
	servletNameMappings []ServletNameMapping
}

// URLPatternMapping 描述 Filter 与 URL 模式之间的映射。
type URLPatternMapping struct {
	dispatchers DispatcherTypes
	matchAfter  bool
	patterns    []string
}

// ServletNameMapping 描述 Filter 与 Servlet 名称之间的映射。
type ServletNameMapping struct {
	dispatchers  DispatcherTypes
	matchAfter   bool
	servletNames []string
}

// Filter 返回注册的 Filter 实例。
func (r *FilterRegistration) Filter() servlet.Filter {
	if r == nil || r.owner == nil {
		return nil
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return r.filter
}

// AddMappingForURLPatterns 为 Filter 增加 URL 模式映射。
func (r *FilterRegistration) AddMappingForURLPatterns(dispatchers DispatcherTypes, matchAfter bool, patterns ...string) error {
	if r == nil || r.owner == nil {
		return ErrNilRegistry
	}
	if err := validateDispatcherTypes(dispatchers); err != nil {
		return err
	}
	normalized := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		if err := validateURLPattern(pattern); err != nil {
			return err
		}
		normalized = appendMissingString(normalized, pattern)
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return err
	}
	mapping := URLPatternMapping{
		dispatchers: normalizeDispatcherTypes(dispatchers),
		matchAfter:  matchAfter,
		patterns:    normalized,
	}
	r.urlPatternMappings = append(r.urlPatternMappings, mapping)
	return nil
}

// AddMappingForServletNames 为 Filter 增加 Servlet 名称映射。
func (r *FilterRegistration) AddMappingForServletNames(dispatchers DispatcherTypes, matchAfter bool, names ...string) error {
	if r == nil || r.owner == nil {
		return ErrNilRegistry
	}
	if err := validateDispatcherTypes(dispatchers); err != nil {
		return err
	}
	normalized := make([]string, 0, len(names))
	for _, name := range names {
		if err := validateName(name); err != nil {
			return err
		}
		normalized = appendMissingString(normalized, name)
	}
	r.owner.mu.Lock()
	defer r.owner.mu.Unlock()
	if err := r.owner.ensureMutableLocked(); err != nil {
		return err
	}
	mapping := ServletNameMapping{
		dispatchers:  normalizeDispatcherTypes(dispatchers),
		matchAfter:   matchAfter,
		servletNames: normalized,
	}
	r.servletNameMappings = append(r.servletNameMappings, mapping)
	return nil
}

// URLPatternMappings 返回 URL 模式映射副本。
func (r *FilterRegistration) URLPatternMappings() []URLPatternMapping {
	if r == nil || r.owner == nil {
		return nil
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return cloneURLPatternMappings(r.urlPatternMappings)
}

// ServletNameMappings 返回 Servlet 名称映射副本。
func (r *FilterRegistration) ServletNameMappings() []ServletNameMapping {
	if r == nil || r.owner == nil {
		return nil
	}
	r.owner.mu.RLock()
	defer r.owner.mu.RUnlock()
	return cloneServletNameMappings(r.servletNameMappings)
}

// DispatcherTypes 返回映射声明的分发类型集合。
func (m URLPatternMapping) DispatcherTypes() DispatcherTypes {
	return m.dispatchers
}

// MatchAfter 表示该映射是否追加在已有映射之后。
func (m URLPatternMapping) MatchAfter() bool {
	return m.matchAfter
}

// URLPatterns 返回 URL 模式副本。
func (m URLPatternMapping) URLPatterns() []string {
	return cloneStrings(m.patterns)
}

// DispatcherTypes 返回映射声明的分发类型集合。
func (m ServletNameMapping) DispatcherTypes() DispatcherTypes {
	return m.dispatchers
}

// MatchAfter 表示该映射是否追加在已有映射之后。
func (m ServletNameMapping) MatchAfter() bool {
	return m.matchAfter
}

// ServletNames 返回 Servlet 名称副本。
func (m ServletNameMapping) ServletNames() []string {
	return cloneStrings(m.servletNames)
}

func cloneURLPatternMappings(src []URLPatternMapping) []URLPatternMapping {
	if len(src) == 0 {
		return nil
	}
	dst := make([]URLPatternMapping, len(src))
	copy(dst, src)
	for i := range dst {
		dst[i].patterns = cloneStrings(dst[i].patterns)
	}
	return dst
}

func cloneServletNameMappings(src []ServletNameMapping) []ServletNameMapping {
	if len(src) == 0 {
		return nil
	}
	dst := make([]ServletNameMapping, len(src))
	copy(dst, src)
	for i := range dst {
		dst[i].servletNames = cloneStrings(dst[i].servletNames)
	}
	return dst
}
