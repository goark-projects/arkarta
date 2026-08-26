package resource

import (
	"context"
	"io"
	"time"
)

// Provider 按 Servlet 资源路径打开静态资源。
type Provider interface {
	Open(ctx context.Context, path string) (Resource, error)
}

// Resource 表示一次打开的静态资源。
type Resource struct {
	path        string
	size        int64
	modTime     time.Time
	contentType string
	etag        string
	body        io.ReadCloser
}

func newResource(path string, size int64, modTime time.Time, contentType, etag string, body io.ReadCloser) Resource {
	return Resource{
		path:        path,
		size:        size,
		modTime:     modTime,
		contentType: contentType,
		etag:        etag,
		body:        body,
	}
}

// Path 返回规整后的 Web 资源路径。
func (r Resource) Path() string {
	return r.path
}

// Size 返回资源字节数。
func (r Resource) Size() int64 {
	return r.size
}

// ModTime 返回资源最后修改时间。
func (r Resource) ModTime() time.Time {
	return r.modTime
}

// ContentType 返回资源媒体类型。
func (r Resource) ContentType() string {
	return r.contentType
}

// ETag 返回资源弱实体标签。
func (r Resource) ETag() string {
	return r.etag
}

// Body 返回资源内容读取器；调用方负责关闭。
func (r Resource) Body() io.ReadCloser {
	return r.body
}
