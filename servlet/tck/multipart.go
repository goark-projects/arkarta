package tck

import (
	"bytes"
	"errors"
	"io"
	stdmultipart "mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/multipart"
)

// MultipartParserFactory 创建 multipart 解析器。
type MultipartParserFactory func(options ...multipart.Option) *multipart.Parser

// RunMultipartParser 执行 Multipart Profile 解析器兼容性测试。
func RunMultipartParser(t *testing.T, factory MultipartParserFactory) {
	t.Helper()
	t.Run("parse_values_and_files", func(t *testing.T) {
		req := newMultipartRequest(t, "field", "value")
		form, err := factory().Parse(req)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		defer form.RemoveAll()
		if form.Value("field") != "value" {
			t.Fatalf("field = %q, want value", form.Value("field"))
		}
		files, ok := form.File("file")
		if !ok || len(files) != 1 || files[0].Filename != "tck.txt" {
			t.Fatalf("files = %#v/%v", files, ok)
		}
	})
	t.Run("body_limit", func(t *testing.T) {
		req := newMultipartRequest(t, "field", "value")
		_, err := factory(multipart.WithMaxBodySize(8)).Parse(req)
		if !errors.Is(err, multipart.ErrBodyTooLarge) {
			t.Fatalf("Parse err = %v, want ErrBodyTooLarge", err)
		}
	})
	t.Run("storage_location_filename_and_cleanup", func(t *testing.T) {
		runMultipartStorageLocationFilenameAndCleanup(t, factory)
	})
}

func newMultipartRequest(t *testing.T, field, value string) *servlet.Request {
	t.Helper()
	return newMultipartRequestWithFile(t, field, value, "tck.txt")
}

func newMultipartRequestWithFile(t *testing.T, field, value, filename string) *servlet.Request {
	t.Helper()
	var body bytes.Buffer
	writer := stdmultipart.NewWriter(&body)
	if err := writer.WriteField(field, value); err != nil {
		t.Fatalf("WriteField failed: %v", err)
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write([]byte("hello")); err != nil {
		t.Fatalf("file write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer close failed: %v", err)
	}
	httpRequest := httptest.NewRequest(http.MethodPost, "/upload", &body)
	httpRequest.Header.Set("Content-Type", writer.FormDataContentType())
	req, err := servlet.NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	return req
}

func runMultipartStorageLocationFilenameAndCleanup(t *testing.T, factory MultipartParserFactory) {
	t.Helper()
	location := t.TempDir()
	req := newMultipartRequestWithFile(t, "field", "value", `..\secret.txt`)
	config := multipart.NewConfig(multipart.WithLocation(location))
	form, err := factory(multipart.WithConfig(config)).Parse(req)
	if err != nil {
		t.Fatalf("Parse with location failed: %v", err)
	}
	if form.Value("field") != "value" {
		t.Fatalf("field = %q, want value", form.Value("field"))
	}
	part, ok := form.Part("file")
	if !ok {
		t.Fatal("Part(file) should exist")
	}
	if part.SubmittedFileName() != "secret.txt" {
		t.Fatalf("submitted filename = %q, want secret.txt", part.SubmittedFileName())
	}
	reader, err := part.Open()
	if err != nil {
		t.Fatalf("Part.Open failed: %v", err)
	}
	data, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil {
		t.Fatalf("read/close part = %v/%v", readErr, closeErr)
	}
	if string(data) != "hello" {
		t.Fatalf("part content = %q, want hello", string(data))
	}
	if err := part.Write(filepath.Join(location, "copy.txt")); err != nil {
		t.Fatalf("Part.Write failed: %v", err)
	}
	if err := part.Delete(); err != nil {
		t.Fatalf("Part.Delete failed: %v", err)
	}
	if _, err := part.Open(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Part.Open after Delete err = %v, want not exist", err)
	}
	if err := form.RemoveAll(); err != nil {
		t.Fatalf("RemoveAll after Part.Delete failed: %v", err)
	}
}
