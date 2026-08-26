package servlet

import "strings"

const (
	// AttributeTempDir 保存应用临时目录。
	AttributeTempDir = "arkarta.servlet.context.tempdir"
)

// WithTempDir 设置应用临时目录。
func WithTempDir(path string) WebAppOption {
	return func(app *WebApp) error {
		path = strings.TrimSpace(path)
		app.tempDir = path
		if path != "" {
			app.attribute[AttributeTempDir] = path
		}
		return nil
	}
}

// TempDir 返回应用临时目录。
func (a *WebApp) TempDir() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tempDir
}
