package multipart

import (
	"bytes"
	"errors"
	stdmultipart "mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"goark.dev/arkarta/servlet"
)

func TestParserParsesValuesAndFiles(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := stdmultipart.NewWriter(&body)
	if err := writer.WriteField("name", "arkarta"); err != nil {
		t.Fatalf("WriteField failed: %v", err)
	}
	file, err := writer.CreateFormFile("artifact", "readme.txt")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := file.Write([]byte("hello")); err != nil {
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

	form, err := NewParser().Parse(req)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer form.RemoveAll()

	if form.Value("name") != "arkarta" {
		t.Fatalf("name = %q, want arkarta", form.Value("name"))
	}
	files, ok := form.File("artifact")
	if !ok || len(files) != 1 || files[0].Filename != "readme.txt" {
		t.Fatalf("files = %#v/%v, want readme.txt", files, ok)
	}
	part, ok := form.Part("artifact")
	if !ok || part.Name() != "artifact" || part.SubmittedFileName() != "readme.txt" || part.Size() != int64(len("hello")) {
		t.Fatalf("part = %#v/%v", part, ok)
	}
	target := filepath.Join(t.TempDir(), "copy.txt")
	if err := part.Write(target); err != nil {
		t.Fatalf("Part.Write failed: %v", err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("copied content = %q, want hello", string(content))
	}
	if bound, ok := Current(req); !ok || bound != form {
		t.Fatalf("bound form = %v/%v, want parsed form", bound, ok)
	}
}

func TestParserRejectsNonMultipartRequest(t *testing.T) {
	t.Parallel()

	httpRequest := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("plain"))
	httpRequest.Header.Set("Content-Type", "text/plain")
	req, err := servlet.NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	_, err = NewParser().Parse(req)
	if !errors.Is(err, ErrNotMultipart) {
		t.Fatalf("Parse err = %v, want ErrNotMultipart", err)
	}
}

func TestParserEnforcesBodyLimit(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := stdmultipart.NewWriter(&body)
	if err := writer.WriteField("name", "arkarta"); err != nil {
		t.Fatalf("WriteField failed: %v", err)
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

	_, err = NewParser(WithMaxBodySize(8)).Parse(req)
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("Parse err = %v, want ErrBodyTooLarge", err)
	}
}

func TestParserEnforcesFileLimit(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := stdmultipart.NewWriter(&body)
	file, err := writer.CreateFormFile("artifact", "readme.txt")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := file.Write([]byte("hello")); err != nil {
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

	config := NewConfig(WithMaxFileSize(4), WithMaxRequestSize(1024))
	_, err = NewParser(WithConfig(config)).Parse(req)
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("Parse err = %v, want ErrFileTooLarge", err)
	}
}

func TestParserStoresFilesInConfiguredLocationAndDeletesThem(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := stdmultipart.NewWriter(&body)
	file, err := writer.CreateFormFile("artifact", "readme.txt")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := file.Write([]byte("hello")); err != nil {
		t.Fatalf("file write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer close failed: %v", err)
	}

	location := t.TempDir()
	httpRequest := httptest.NewRequest(http.MethodPost, "/upload", &body)
	httpRequest.Header.Set("Content-Type", writer.FormDataContentType())
	req, err := servlet.NewRequest(httpRequest)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	form, err := NewParser(WithConfig(NewConfig(WithLocation(location)))).Parse(req)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	part, ok := form.Part("artifact")
	if !ok {
		t.Fatal("part should exist")
	}
	if part.path == "" || filepath.Dir(part.path) != location {
		t.Fatalf("part path = %q, want under %q", part.path, location)
	}
	if _, err := os.Stat(part.path); err != nil {
		t.Fatalf("temp file stat failed: %v", err)
	}
	if err := part.Delete(); err != nil {
		t.Fatalf("Part.Delete failed: %v", err)
	}
	if _, err := os.Stat(part.path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file err = %v, want not exist", err)
	}
}

func TestParserNormalizesSubmittedFileName(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := stdmultipart.NewWriter(&body)
	file, err := writer.CreateFormFile("artifact", `C:\fake\..\readme.txt`)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := file.Write([]byte("hello")); err != nil {
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
	form, err := NewParser(WithConfig(NewConfig(WithLocation(t.TempDir())))).Parse(req)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	defer form.RemoveAll()

	part, ok := form.Part("artifact")
	if !ok || part.SubmittedFileName() != "readme.txt" {
		t.Fatalf("submitted filename = %q/%v, want readme.txt/true", part.SubmittedFileName(), ok)
	}
}

func TestParseRequestReusesBoundForm(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := stdmultipart.NewWriter(&body)
	if err := writer.WriteField("name", "arkarta"); err != nil {
		t.Fatalf("WriteField failed: %v", err)
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

	first, err := ParseRequest(req, NewParser())
	if err != nil {
		t.Fatalf("ParseRequest first failed: %v", err)
	}
	second, err := ParseRequest(req, NewParser(WithMaxBodySize(1)))
	if err != nil {
		t.Fatalf("ParseRequest second failed: %v", err)
	}
	if first != second {
		t.Fatal("ParseRequest should reuse request-bound form")
	}
}
