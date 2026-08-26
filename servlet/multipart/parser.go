package multipart

import (
	"bytes"
	"errors"
	"io"
	"mime"
	stdmultipart "mime/multipart"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"strings"

	"goark.dev/arkarta/servlet"
)

// Parser 解析 Servlet multipart/form-data 请求。
type Parser struct {
	maxMemory   int64
	maxBodySize int64
	maxFileSize int64
	location    string
}

// NewParser 创建 Multipart 解析器。
func NewParser(options ...Option) *Parser {
	parser := &Parser{
		maxMemory:   defaultMaxMemory,
		maxBodySize: -1,
		maxFileSize: -1,
	}
	for _, option := range options {
		if option != nil {
			option(parser)
		}
	}
	return parser
}

// Parse 解析请求体。
func (p *Parser) Parse(req *servlet.Request) (*Form, error) {
	if req == nil || req.HTTPRequest() == nil {
		return nil, servlet.ErrNilHTTPRequest
	}
	if !isMultipart(req) {
		return nil, ErrNotMultipart
	}

	httpRequest := req.HTTPRequest()
	if p.maxBodySize >= 0 {
		httpRequest.Body = &limitReadCloser{
			reader:    httpRequest.Body,
			remaining: p.maxBodySize,
		}
	}
	if p.location != "" {
		return p.parseWithLocation(req)
	}
	if err := httpRequest.ParseMultipartForm(p.maxMemory); err != nil {
		if errors.Is(err, ErrBodyTooLarge) {
			return nil, err
		}
		return nil, err
	}
	form := &Form{form: httpRequest.MultipartForm}
	if err := p.validateFiles(form); err != nil {
		_ = form.RemoveAll()
		return nil, err
	}
	req.SetAttribute(AttributeForm, form)
	return form, nil
}

func (p *Parser) validateFiles(form *Form) error {
	if p.maxFileSize < 0 {
		return nil
	}
	for _, files := range form.Files() {
		for _, file := range files {
			if file.Size > p.maxFileSize {
				return ErrFileTooLarge
			}
		}
	}
	return nil
}

func (p *Parser) parseWithLocation(req *servlet.Request) (*Form, error) {
	if err := os.MkdirAll(p.location, 0o700); err != nil {
		return nil, err
	}
	reader, err := req.HTTPRequest().MultipartReader()
	if err != nil {
		return nil, err
	}
	form := &Form{values: make(map[string][]string)}
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			_ = form.RemoveAll()
			return nil, err
		}
		if part.FormName() == "" {
			_ = part.Close()
			continue
		}
		if part.FileName() == "" {
			name := part.FormName()
			value, err := readValuePart(part, p.maxMemory)
			if err != nil {
				_ = form.RemoveAll()
				return nil, err
			}
			form.values.Add(name, value)
			continue
		}
		stored, err := p.storeFilePart(part)
		if err != nil {
			_ = form.RemoveAll()
			return nil, err
		}
		form.parts = append(form.parts, stored)
	}
	req.SetAttribute(AttributeForm, form)
	return form, nil
}

func readValuePart(part interface {
	io.Reader
	Close() error
}, maxMemory int64) (string, error) {
	defer part.Close()
	var buffer bytes.Buffer
	if maxMemory > 0 {
		if _, err := io.CopyN(&buffer, part, maxMemory+1); err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		if int64(buffer.Len()) > maxMemory {
			return "", ErrBodyTooLarge
		}
		return buffer.String(), nil
	}
	_, err := io.Copy(&buffer, part)
	return buffer.String(), err
}

func (p *Parser) storeFilePart(part *stdmultipart.Part) (Part, error) {
	defer part.Close()
	file, err := os.CreateTemp(p.location, "arkarta-multipart-*")
	if err != nil {
		return Part{}, err
	}
	path := file.Name()
	size, copyErr := copyFilePart(file, part, p.maxFileSize)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(path)
		return Part{}, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return Part{}, closeErr
	}
	return Part{
		name:     part.FormName(),
		header:   cloneMIMEHeader(textproto.MIMEHeader(part.Header)),
		size:     size,
		path:     path,
		filename: cleanSubmittedFileName(part.FileName()),
	}, nil
}

func copyFilePart(dst io.Writer, src io.Reader, maxFileSize int64) (int64, error) {
	if maxFileSize < 0 {
		return io.Copy(dst, src)
	}
	limited := &limitReader{reader: src, remaining: maxFileSize}
	n, err := io.Copy(dst, limited)
	if errors.Is(err, ErrFileTooLarge) {
		return n, err
	}
	return n, err
}

type limitReader struct {
	reader    io.Reader
	remaining int64
}

func (r *limitReader) Read(buffer []byte) (int, error) {
	if r.remaining < 0 {
		return r.reader.Read(buffer)
	}
	if r.remaining == 0 {
		var probe [1]byte
		n, err := r.reader.Read(probe[:])
		if n > 0 {
			return 0, ErrFileTooLarge
		}
		return 0, err
	}
	if int64(len(buffer)) > r.remaining+1 {
		buffer = buffer[:r.remaining+1]
	}
	n, err := r.reader.Read(buffer)
	if int64(n) > r.remaining {
		allowed := int(r.remaining)
		r.remaining = 0
		return allowed, ErrFileTooLarge
	}
	r.remaining -= int64(n)
	return n, err
}

func cleanSubmittedFileName(filename string) string {
	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = path.Clean("/" + filename)
	filename = path.Base(filename)
	if filename == "." || filename == "/" {
		return ""
	}
	return filepath.Base(filename)
}

func isMultipart(req *servlet.Request) bool {
	contentType := req.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.ToLower(mediaType), "multipart/")
}
