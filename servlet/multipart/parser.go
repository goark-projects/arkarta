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
}

// NewParser 创建 Multipart 解析器。
func NewParser(options ...Option) *Parser {
	parser := &Parser{
		maxMemory:   defaultMaxMemory,
		maxBodySize: -1,
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
	return &Form{form: httpRequest.MultipartForm}, nil
}

func isMultipart(req *servlet.Request) bool {
	contentType := req.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.ToLower(mediaType), "multipart/")
}
