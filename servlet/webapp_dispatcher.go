package servlet

// DispatcherProvider 按路径或名称提供请求分发器。
type DispatcherProvider interface {
	RequestDispatcher(path string) (RequestDispatcher, error)
	NamedDispatcher(name string) (RequestDispatcher, error)
}

// WithDispatcherProvider 设置请求分发器提供者。
func WithDispatcherProvider(provider DispatcherProvider) WebAppOption {
	return func(app *WebApp) error {
		app.dispatcherProvider = provider
		return nil
	}
}

// SetDispatcherProvider 设置请求分发器提供者。
func (a *WebApp) SetDispatcherProvider(provider DispatcherProvider) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.dispatcherProvider = provider
}

// RequestDispatcher 返回指定路径的请求分发器。
func (a *WebApp) RequestDispatcher(path string) (RequestDispatcher, error) {
	a.mu.RLock()
	provider := a.dispatcherProvider
	a.mu.RUnlock()
	if provider == nil {
		return nil, ErrNilRouter
	}
	return provider.RequestDispatcher(path)
}

// NamedDispatcher 返回指定 Servlet 名称的请求分发器。
func (a *WebApp) NamedDispatcher(name string) (RequestDispatcher, error) {
	a.mu.RLock()
	provider := a.dispatcherProvider
	a.mu.RUnlock()
	if provider == nil {
		return nil, ErrDispatcherTargetNotFound
	}
	return provider.NamedDispatcher(name)
}
