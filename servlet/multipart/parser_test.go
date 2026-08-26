package multipart

import (
	"bytes"
	"errors"
	stdmultipart "mime/multipart"
	"net/http"
	"net/http/httptest"
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
