package multipart

import "goark.dev/arkarta/servlet"

const (
	// AttributeForm 保存当前请求已解析的 multipart 表单。
	AttributeForm = "arkarta.servlet.multipart.form"
)

// Current 返回当前请求已绑定的 multipart 表单。
func Current(req *servlet.Request) (*Form, bool) {
	if req == nil {
		return nil, false
	}
	value, ok := req.Attribute(AttributeForm)
	if !ok {
		return nil, false
	}
	form, ok := value.(*Form)
	return form, ok && form != nil
}

// ParseRequest 解析并绑定当前请求 multipart 表单。
func ParseRequest(req *servlet.Request, parser *Parser) (*Form, error) {
	if form, ok := Current(req); ok {
		return form, nil
	}
	if parser == nil {
		parser = NewParser()
	}
	return parser.Parse(req)
}

// RequestPart 返回当前请求指定字段的第一个文件段。
func RequestPart(req *servlet.Request, name string, parser *Parser) (Part, bool, error) {
	form, err := ParseRequest(req, parser)
	if err != nil {
		return Part{}, false, err
	}
	part, ok := form.Part(name)
	return part, ok, nil
}

// Parts 返回当前请求的所有文件段。
func Parts(req *servlet.Request, parser *Parser) ([]Part, error) {
	form, err := ParseRequest(req, parser)
	if err != nil {
		return nil, err
	}
	return form.Parts(), nil
}
