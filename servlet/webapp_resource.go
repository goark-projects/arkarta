package servlet

import (
	"context"
	"errors"
	"io/fs"
	"path"
	"strings"
)

// WithResourceFS 设置应用上下文资源文件系统。
func WithResourceFS(root fs.FS) WebAppOption {
	return func(app *WebApp) error {
		if root == nil {
			return ErrInvalidWebAppConfig
		}
		app.resourceFS = root
		return nil
	}
}

// OpenResource 按 Servlet 资源路径打开应用资源。
func (a *WebApp) OpenResource(ctx context.Context, value string) (fs.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	name, err := cleanWebAppResourceName(value)
	if err != nil {
		return nil, err
	}
	a.mu.RLock()
	root := a.resourceFS
	a.mu.RUnlock()
	if root == nil {
		return nil, fs.ErrNotExist
	}
	file, err := root.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}
	return file, nil
}

// ResourceExists 判断应用资源是否存在。
func (a *WebApp) ResourceExists(ctx context.Context, value string) (bool, error) {
	file, err := a.OpenResource(ctx, value)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	_ = file.Close()
	return true, nil
}

func cleanWebAppResourceName(value string) (string, error) {
	if value == "" {
		value = "/"
	}
	if strings.ContainsAny(value, "\x00\\") {
		return "", ErrInvalidWebAppConfig
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == ".." {
			return "", ErrInvalidWebAppConfig
		}
	}
	clean := strings.TrimPrefix(path.Clean(value), "/")
	if clean == "" {
		return ".", nil
	}
	if !fs.ValidPath(clean) {
		return "", ErrInvalidWebAppConfig
	}
	return clean, nil
}
