package nativeio

import "errors"

// ErrNilWriter 表示发送目标为空。
var ErrNilWriter = errors.New("arkarta/servlet/nativeio: writer is nil")

// ErrNilSource 表示文件区段缺少可随机读取的数据源。
var ErrNilSource = errors.New("arkarta/servlet/nativeio: source is nil")

// ErrInvalidRegion 表示文件区段范围非法。
var ErrInvalidRegion = errors.New("arkarta/servlet/nativeio: invalid file region")
