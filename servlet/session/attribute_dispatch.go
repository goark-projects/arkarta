package session

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

func (m *MemoryManager) fireAttributeReplaced(target Session, name string, value, oldValue any) {
	event := AttributeEvent{Session: target, Name: name, Value: value, OldValue: oldValue}
	for _, listener := range m.attributeListenerSnapshot() {
		listener.AttributeReplaced(event)
	}
}

func (m *MemoryManager) fireAttributeRemoved(target Session, name string, oldValue any) {
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
