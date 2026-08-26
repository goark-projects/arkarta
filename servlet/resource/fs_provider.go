package resource

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"path"
	"strconv"
)

// ContentTypeFunc 根据资源路径返回媒体类型。
type ContentTypeFunc func(path string) string

// FSProviderOption 定制文件系统资源提供者。
type FSProviderOption func(*FSProvider)

// WithContentTypeFunc 设置媒体类型探测函数。
func WithContentTypeFunc(fn ContentTypeFunc) FSProviderOption {
	return func(provider *FSProvider) {
		if fn != nil {
			provider.contentType = fn
		}
	}
}

// FSProvider 从标准 fs.FS 读取静态资源。
type FSProvider struct {
	root        fs.FS
	contentType ContentTypeFunc
}

// NewFSProvider 创建 fs.FS 资源提供者。
func NewFSProvider(root fs.FS, options ...FSProviderOption) (*FSProvider, error) {
	if root == nil {
		return nil, ErrNilFileSystem
	}
	provider := &FSProvider{
		root:        root,
		contentType: detectContentType,
	}
	for _, option := range options {
		if option != nil {
			option(provider)
		}
	}
	return provider, nil
}

// Open 打开指定静态资源。
func (p *FSProvider) Open(ctx context.Context, value string) (Resource, error) {
	if err := ctx.Err(); err != nil {
		return Resource{}, err
	}
	clean, name, err := cleanResourcePath(value)
	if err != nil {
		return Resource{}, err
	}
	file, err := p.root.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Resource{}, ErrNotFound
		}
		return Resource{}, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return Resource{}, err
	}
	if info.IsDir() {
		_ = file.Close()
		return Resource{}, ErrDirectory
	}
	return newResource(
		clean,
		info.Size(),
		info.ModTime(),
		p.contentType(clean),
		weakETag(info.Size(), info.ModTime().UnixNano()),
		file,
	), nil
}

func detectContentType(value string) string {
	if contentType := mime.TypeByExtension(path.Ext(value)); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}

func weakETag(size, modUnixNano int64) string {
	return fmt.Sprintf(`W/"%s-%s"`, strconv.FormatInt(size, 16), strconv.FormatInt(modUnixNano, 16))
}
