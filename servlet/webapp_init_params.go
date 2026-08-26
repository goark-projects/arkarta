package servlet

import (
	"fmt"
	"sort"
	"strings"
)

// SetInitParam 在应用启动前设置初始化参数；返回 false 表示名称已存在。
func (a *WebApp) SetInitParam(name, value string) (bool, error) {
	if strings.TrimSpace(name) == "" {
		return false, ErrInvalidWebAppConfig
	}
	if a == nil {
		return false, ErrInvalidWebAppConfig
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.ensureInitParamMutableLocked(); err != nil {
		return false, err
	}
	if _, exists := a.initParam[name]; exists {
		return false, nil
	}
	a.initParam[name] = value
	return true, nil
}

// SetInitParams 批量设置初始化参数；存在冲突时不写入任何参数。
func (a *WebApp) SetInitParams(params map[string]string) ([]string, error) {
	for name := range params {
		if strings.TrimSpace(name) == "" {
			return nil, ErrInvalidWebAppConfig
		}
	}
	if a == nil {
		return nil, ErrInvalidWebAppConfig
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.ensureInitParamMutableLocked(); err != nil {
		return nil, err
	}
	conflicts := make([]string, 0)
	for name := range params {
		if _, exists := a.initParam[name]; exists {
			conflicts = append(conflicts, name)
		}
	}
	sort.Strings(conflicts)
	if len(conflicts) > 0 {
		return conflicts, nil
	}
	for name, value := range params {
		a.initParam[name] = value
	}
	return nil, nil
}

func (a *WebApp) ensureInitParamMutableLocked() error {
	switch a.state {
	case WebAppStateNew, WebAppStateInitialized:
		return nil
	default:
		return fmt.Errorf("%w: cannot set init parameter from %v", ErrInvalidWebAppState, a.state)
	}
}
