package multipart

import (
	"io"
	stdmultipart "mime/multipart"
	"net/textproto"
	"os"
)

// Part 表示 multipart/form-data 中的一个文件段。
type Part struct {
	name    string
	file    *stdmultipart.FileHeader
	cleanup func() error
}

// Name 返回表单字段名。
func (p Part) Name() string {
	return p.name
}

// SubmittedFileName 返回客户端提交的文件名。
func (p Part) SubmittedFileName() string {
	if p.file == nil {
		return ""
	}
	return p.file.Filename
}

// Header 返回文件段头部副本。
func (p Part) Header() textproto.MIMEHeader {
	if p.file == nil {
		return textproto.MIMEHeader{}
	}
	result := make(textproto.MIMEHeader, len(p.file.Header))
	for name, values := range p.file.Header {
		result[name] = append([]string(nil), values...)
	}
	return result
}

// Size 返回文件段大小。
func (p Part) Size() int64 {
	if p.file == nil {
		return 0
	}
	return p.file.Size
}

// Open 打开文件段内容流。
func (p Part) Open() (stdmultipart.File, error) {
	if p.file == nil {
		return nil, os.ErrNotExist
	}
	return p.file.Open()
}

// Write 将文件段内容写入目标文件。
func (p Part) Write(path string) error {
	src, err := p.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// Delete 清理该 Part 所属表单的临时文件。
func (p Part) Delete() error {
	if p.cleanup == nil {
		return nil
	}
	return p.cleanup()
}
