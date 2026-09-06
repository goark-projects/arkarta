package session

// AttributeEvent 表示会话属性变更事件。
type AttributeEvent struct {
	Session  Session
	Name     string
	Value    any
	OldValue any
}

// AttributeListener 监听会话属性变更。
type AttributeListener interface {
	AttributeAdded(event AttributeEvent)
	AttributeReplaced(event AttributeEvent)
	AttributeRemoved(event AttributeEvent)
}

// AttributeListenerFunc 将函数组适配为 AttributeListener。
type AttributeListenerFunc struct {
	Added    func(event AttributeEvent)
	Replaced func(event AttributeEvent)
	Removed  func(event AttributeEvent)
}

// AttributeAdded 触发属性新增回调。
func (f AttributeListenerFunc) AttributeAdded(event AttributeEvent) {
	if f.Added != nil {
		f.Added(event)
	}
}

// AttributeReplaced 触发属性替换回调。
func (f AttributeListenerFunc) AttributeReplaced(event AttributeEvent) {
	if f.Replaced != nil {
		f.Replaced(event)
	}
}

// AttributeRemoved 触发属性移除回调。
func (f AttributeListenerFunc) AttributeRemoved(event AttributeEvent) {
	if f.Removed != nil {
		f.Removed(event)
	}
}

// BindingEvent 表示属性值绑定或解绑事件。
type BindingEvent struct {
	Session Session
	Name    string
	Value   any
}

// BindingListener 由会话属性值实现，用于感知自身绑定状态。
type BindingListener interface {
	ValueBound(event BindingEvent)
	ValueUnbound(event BindingEvent)
}

// ActivationEvent 表示会话激活或钝化事件。
type ActivationEvent struct {
	Session Session
}

// ActivationListener 由需要感知分布式会话迁移的值实现。
type ActivationListener interface {
	SessionWillPassivate(event ActivationEvent)
	SessionDidActivate(event ActivationEvent)
}

func (m *MemoryManager) attributeListenerSnapshot() []AttributeListener {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]AttributeListener, len(m.attributeListeners))
	copy(result, m.attributeListeners)
	return result
}

func (m *MemoryManager) fireAttributeAdded(target Session, name string, value any) {
	event := AttributeEvent{Session: target, Name: name, Value: value}
	for _, listener := range m.attributeListenerSnapshot() {
		listener.AttributeAdded(event)
	}
}

func (m *MemoryManager) fireAttributeReplaced(
	target Session,
	name string,
	value, oldValue any,
) {
	event := AttributeEvent{
		Session:  target,
		Name:     name,
		Value:    value,
		OldValue: oldValue,
	}
	for _, listener := range m.attributeListenerSnapshot() {
		listener.AttributeReplaced(event)
	}
}

func (m *MemoryManager) fireAttributeRemoved(
	target Session,
	name string,
	oldValue any,
) {
	event := AttributeEvent{Session: target, Name: name, OldValue: oldValue}
	for _, listener := range m.attributeListenerSnapshot() {
		listener.AttributeRemoved(event)
	}
}

func (m *MemoryManager) fireAttributesRemoved(target Session, values map[string]any) {
	for name, value := range values {
		fireValueUnbound(target, name, value)
		m.fireAttributeRemoved(target, name, value)
	}
}

func fireValueBound(target Session, name string, value any) {
	listener, ok := value.(BindingListener)
	if !ok || listener == nil {
		return
	}
	listener.ValueBound(BindingEvent{Session: target, Name: name, Value: value})
}

func fireValueUnbound(target Session, name string, value any) {
	listener, ok := value.(BindingListener)
	if !ok || listener == nil {
		return
	}
	listener.ValueUnbound(BindingEvent{Session: target, Name: name, Value: value})
}

func fireSessionWillPassivate(target Session, values map[string]any) {
	event := ActivationEvent{Session: target}
	for _, value := range values {
		listener, ok := value.(ActivationListener)
		if ok && listener != nil {
			listener.SessionWillPassivate(event)
		}
	}
}

func fireSessionDidActivate(target Session, values map[string]any) {
	event := ActivationEvent{Session: target}
	for _, value := range values {
		listener, ok := value.(ActivationListener)
		if ok && listener != nil {
			listener.SessionDidActivate(event)
		}
	}
}
