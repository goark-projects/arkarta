package servlet

import (
	"context"
	"log/slog"
)

// WithLogger 设置应用上下文日志入口。
func WithLogger(logger *slog.Logger) WebAppOption {
	return func(app *WebApp) error {
		if logger == nil {
			return ErrInvalidWebAppConfig
		}
		app.logger = logger
		return nil
	}
}

// Logger 返回应用上下文日志入口。
func (a *WebApp) Logger() *slog.Logger {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.logger
}

// Log 写出应用上下文日志。
func (a *WebApp) Log(ctx context.Context, message string, args ...any) {
	logger := a.Logger()
	if logger == nil {
		return
	}
	logger.InfoContext(ctx, message, args...)
}
