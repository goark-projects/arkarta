package resource

import "errors"

// ErrNilProvider 表示 default servlet 缺少资源提供者。
var ErrNilProvider = errors.New("arkarta/servlet/resource: provider is nil")

// ErrNilFileSystem 表示文件系统为空。
var ErrNilFileSystem = errors.New("arkarta/servlet/resource: file system is nil")

// ErrInvalidPath 表示资源路径非法。
var ErrInvalidPath = errors.New("arkarta/servlet/resource: invalid resource path")

// ErrNotFound 表示资源不存在。
var ErrNotFound = errors.New("arkarta/servlet/resource: resource not found")

// ErrDirectory 表示请求目标是目录而不是普通资源。
var ErrDirectory = errors.New("arkarta/servlet/resource: resource is directory")
