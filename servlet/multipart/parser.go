package multipart

import (
	"errors"
	"mime"
	"strings"

	"goark.dev/arkarta/servlet"
)

// Parser 解析 Servlet multipart/form-data 请求。
type Parser struct {
	maxMemory   int64
	maxBodySize int64
	maxFileSize int64
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

func isMultipart(req *servlet.Request) bool {
	contentType := req.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.ToLower(mediaType), "multipart/")
}
