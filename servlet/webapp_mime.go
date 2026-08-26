package servlet

import (
	"mime"
	"path"
	"strings"
)

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
