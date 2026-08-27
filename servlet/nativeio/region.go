package nativeio

import "io"

// FileRegion 描述待发送文件的稳定字节区段。
type FileRegion struct {
	source io.ReaderAt
	offset int64
	count  int64
}

// NewFileRegion 创建文件区段。
func NewFileRegion(source io.ReaderAt, offset, count int64) (FileRegion, error) {
	if source == nil {
		return FileRegion{}, ErrNilSource
	}
	if offset < 0 || count < 0 {
		return FileRegion{}, ErrInvalidRegion
	}
	return FileRegion{
		source: source,
		offset: offset,
		count:  count,
	}, nil
}

// Source 返回区段数据源。
func (r FileRegion) Source() io.ReaderAt {
	return r.source
}

// Offset 返回发送起始偏移。
func (r FileRegion) Offset() int64 {
	return r.offset
}

// Count 返回最多发送的字节数。
func (r FileRegion) Count() int64 {
	return r.count
}

func (r FileRegion) validate() error {
	if r.source == nil {
		return ErrNilSource
	}
	if r.offset < 0 || r.count < 0 {
		return ErrInvalidRegion
	}
	return nil
}
