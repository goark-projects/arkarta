package servlet

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"mime"
	"path"
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

// WithMimeType 设置扩展名到媒体类型的映射。
func WithMimeType(extension, contentType string) WebAppOption {
	return func(app *WebApp) error {
		extension, err := normalizeMimeExtension(extension)
		if err != nil {
			return err
		}
		contentType = strings.TrimSpace(contentType)
		if _, _, err := mime.ParseMediaType(contentType); err != nil {
			return ErrInvalidWebAppConfig
		}
		app.mimeTypes[extension] = contentType
		return nil
	}
}

// MimeType 返回指定文件名对应的媒体类型。
func (a *WebApp) MimeType(file string) string {
	extension := strings.ToLower(path.Ext(file))
	if extension == "" {
		return ""
	}
	a.mu.RLock()
	contentType := a.mimeTypes[extension]
	a.mu.RUnlock()
	if contentType != "" {
		return contentType
	}
	return mime.TypeByExtension(extension)
}

// MimeMappings 返回当前 MIME 映射副本。
func (a *WebApp) MimeMappings() map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return cloneStringMap(a.mimeTypes)
}

func defaultMimeMappings() map[string]string {
	return map[string]string{
		".css":  "text/css; charset=utf-8",
		".gif":  "image/gif",
		".htm":  "text/html; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".jpeg": "image/jpeg",
		".jpg":  "image/jpeg",
		".js":   "text/javascript; charset=utf-8",
		".json": "application/json",
		".png":  "image/png",
		".svg":  "image/svg+xml",
		".txt":  "text/plain; charset=utf-8",
		".wasm": "application/wasm",
		".xml":  "application/xml",
	}
}

func normalizeMimeExtension(extension string) (string, error) {
	extension = strings.TrimSpace(strings.ToLower(extension))
	if extension == "" {
		return "", ErrInvalidWebAppConfig
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	if strings.ContainsAny(extension, `/\`) || extension == "." {
		return "", ErrInvalidWebAppConfig
	}
	return extension, nil
}

// ResourcePaths 返回指定资源目录下的直接子路径。
func (a *WebApp) ResourcePaths(ctx context.Context, value string) ([]string, error) {
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
	entries, err := fs.ReadDir(root, name)
	if err != nil {
		return nil, err
	}
	base := resourcePathPrefix(value)
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		item := path.Join(base, entry.Name())
		if entry.IsDir() {
			item += "/"
		}
		if !strings.HasPrefix(item, "/") {
			item = "/" + item
		}
		result = append(result, item)
	}
	sort.Strings(result)
	return result, nil
}

func resourcePathPrefix(value string) string {
	if value == "" || value == "/" {
		return "/"
	}
	value = "/" + strings.Trim(value, "/")
	return value
}

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

const (
	// ServletSpecMajorVersion 是 Arkarta Servlet 当前对齐的 Jakarta Servlet 主版本。
	ServletSpecMajorVersion = 6
	// ServletSpecMinorVersion 是 Arkarta Servlet 当前对齐的 Jakarta Servlet 次版本。
	ServletSpecMinorVersion = 1
	// ArkartaServletMajorVersion 是 Arkarta Servlet 标准主版本。
	ArkartaServletMajorVersion = 1
	// ArkartaServletMinorVersion 是 Arkarta Servlet 标准次版本。
	ArkartaServletMinorVersion = 0
)

// ServletContext 是 WebApp 的 Servlet 语义别名。
