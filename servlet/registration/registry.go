package registration

import (
	"strings"
	"sync"

	"goark.dev/arkarta/servlet"
)

// Registry 保存一个 Web 应用初始化阶段的动态注册元数据。
type Registry struct {
	mu sync.RWMutex

	frozen bool

	servlets        map[string]*ServletRegistration
	servletOrder    []string
	servletMappings map[string]string

	filters     map[string]*FilterRegistration
	filterOrder []string

	listeners []*ListenerRegistration
}

// NewRegistry 创建空注册表。
func NewRegistry() *Registry {
	return &Registry{
		servlets:        make(map[string]*ServletRegistration),
		servletMappings: make(map[string]string),
		filters:         make(map[string]*FilterRegistration),
	}
}

// AddServlet 注册 Servlet 实例。
func (r *Registry) AddServlet(name string, target servlet.Handler) (*ServletRegistration, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if isNil(target) {
		return nil, ErrNilServlet
	}
	if r == nil {
		return nil, ErrNilRegistry
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.ensureMutableLocked(); err != nil {
		return nil, err
	}
	if _, exists := r.servlets[name]; exists {
		return nil, ErrDuplicateRegistration
	}
	item := &ServletRegistration{
		commonRegistration: newCommonRegistration(r, name, target),
		handler:            target,
	}
	r.servlets[name] = item
	r.servletOrder = append(r.servletOrder, name)
	return item, nil
}

// AddFilter 注册 Filter 实例。
func (r *Registry) AddFilter(name string, target servlet.Filter) (*FilterRegistration, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if isNil(target) {
		return nil, ErrNilFilter
	}
	if r == nil {
		return nil, ErrNilRegistry
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.ensureMutableLocked(); err != nil {
		return nil, err
	}
	if _, exists := r.filters[name]; exists {
		return nil, ErrDuplicateRegistration
	}
	item := &FilterRegistration{
		commonRegistration: newCommonRegistration(r, name, target),
		filter:             target,
	}
	r.filters[name] = item
	r.filterOrder = append(r.filterOrder, name)
	return item, nil
}

// Servlet 按名称查找 Servlet 注册项。
func (r *Registry) Servlet(name string) (*ServletRegistration, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.servlets[name]
	return item, ok
}

// Filter 按名称查找 Filter 注册项。
func (r *Registry) Filter(name string) (*FilterRegistration, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.filters[name]
	return item, ok
}

// Frozen 表示注册表是否已经冻结。
func (r *Registry) Frozen() bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.frozen
}

// Freeze 冻结注册表并返回不可变快照。
func (r *Registry) Freeze() (Snapshot, error) {
	if r == nil {
		return Snapshot{}, ErrNilRegistry
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frozen = true
	return r.snapshotLocked(), nil
}

// Snapshot 返回当前注册表快照，不改变冻结状态。
func (r *Registry) Snapshot() Snapshot {
	if r == nil {
		return Snapshot{}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.snapshotLocked()
}

func (r *Registry) ensureMutableLocked() error {
	if r.frozen {
		return ErrRegistryFrozen
	}
	return nil
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidName
	}
	return nil
}
