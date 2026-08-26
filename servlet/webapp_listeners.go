package servlet

import "fmt"

// AddContextListener 在应用初始化前追加上下文监听器。
func (a *WebApp) AddContextListener(listener ContextListener) error {
	if listener == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state != WebAppStateNew {
		return fmt.Errorf("%w: cannot add context listener from %v", ErrInvalidWebAppState, a.state)
	}
	a.contextListeners = append(a.contextListeners, listener)
	return nil
}

// AddRequestListener 在应用初始化前追加请求生命周期监听器。
func (a *WebApp) AddRequestListener(listener RequestListener) error {
	if listener == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state != WebAppStateNew {
		return fmt.Errorf("%w: cannot add request listener from %v", ErrInvalidWebAppState, a.state)
	}
	a.requestListeners = append(a.requestListeners, listener)
	return nil
}
