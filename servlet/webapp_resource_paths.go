package servlet

import (
	"context"
	"io/fs"
	"path"
	"sort"
	"strings"
)

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
