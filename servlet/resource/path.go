package resource

import (
	"io/fs"
	"path"
	"strings"
)

// CleanPath 将请求路径规整为以 / 开头的 Web 资源路径。
func CleanPath(value string) (string, error) {
	clean, _, err := cleanResourcePath(value)
	return clean, err
}

func cleanResourcePath(value string) (string, string, error) {
	if value == "" {
		value = "/"
	}
	if strings.ContainsAny(value, "\x00\\") {
		return "", "", ErrInvalidPath
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == ".." {
			return "", "", ErrInvalidPath
		}
	}
	clean := path.Clean(value)
	if clean == "." {
		clean = "/"
	}
	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}
	name := strings.TrimPrefix(clean, "/")
	if name == "" {
		return clean, ".", nil
	}
	if !fs.ValidPath(name) {
		return "", "", ErrInvalidPath
	}
	return clean, name, nil
}
