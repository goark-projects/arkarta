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
