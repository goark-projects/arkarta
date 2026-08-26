package multipart

import (
	stdmultipart "mime/multipart"
	"net/url"
	"sort"
)

// Form 表示已解析的 multipart 表单。
type Form struct {
	form *stdmultipart.Form
}

// Value 返回指定表单字段的第一个值。
func (f *Form) Value(name string) string {
	if f == nil || f.form == nil {
		return ""
	}
	values := f.form.Value[name]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Values 返回普通表单字段副本。
func (f *Form) Values() url.Values {
	if f == nil || f.form == nil {
		return url.Values{}
	}
	return cloneValues(f.form.Value)
}

// File 返回指定文件字段。
func (f *Form) File(name string) ([]*stdmultipart.FileHeader, bool) {
	if f == nil || f.form == nil {
		return nil, false
	}
	files, ok := f.form.File[name]
	return cloneFiles(files), ok
}

// Part 返回指定字段的第一个文件段。
func (f *Form) Part(name string) (Part, bool) {
	files, ok := f.File(name)
	if !ok || len(files) == 0 {
		return Part{}, false
	}
	return Part{name: name, file: files[0], cleanup: f.RemoveAll}, true
}

// Parts 返回所有文件段。
func (f *Form) Parts() []Part {
	files := f.Files()
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]Part, 0)
	for _, name := range names {
		list := files[name]
		for _, file := range list {
			parts = append(parts, Part{name: name, file: file, cleanup: f.RemoveAll})
		}
	}
	return parts
}

// Files 返回文件字段副本。
func (f *Form) Files() map[string][]*stdmultipart.FileHeader {
	if f == nil || f.form == nil {
		return map[string][]*stdmultipart.FileHeader{}
	}
	result := make(map[string][]*stdmultipart.FileHeader, len(f.form.File))
	for name, files := range f.form.File {
		result[name] = cloneFiles(files)
	}
	return result
}

// RemoveAll 清理解析过程中产生的临时文件。
func (f *Form) RemoveAll() error {
	if f == nil || f.form == nil {
		return nil
	}
	return f.form.RemoveAll()
}

func cloneValues(src map[string][]string) url.Values {
	result := make(url.Values, len(src))
	for key, values := range src {
		result[key] = append([]string(nil), values...)
	}
	return result
}

func cloneFiles(src []*stdmultipart.FileHeader) []*stdmultipart.FileHeader {
	if len(src) == 0 {
		return nil
	}
	result := make([]*stdmultipart.FileHeader, len(src))
	copy(result, src)
	return result
}
