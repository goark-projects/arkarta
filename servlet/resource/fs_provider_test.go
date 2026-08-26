package resource

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"
)

func TestFSProviderOpensStaticResource(t *testing.T) {
	t.Parallel()

	modTime := time.Date(2026, 8, 26, 10, 0, 0, 0, time.UTC)
	provider, err := NewFSProvider(fstest.MapFS{
		"assets/app.json": &fstest.MapFile{Data: []byte(`{"ok":true}`), ModTime: modTime},
	})
	if err != nil {
		t.Fatalf("NewFSProvider failed: %v", err)
	}

	item, err := provider.Open(context.Background(), "/assets/app.json")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer item.Body().Close()
	data, err := io.ReadAll(item.Body())
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if string(data) != `{"ok":true}` {
		t.Fatalf("body = %q, want json", string(data))
	}
	if item.Path() != "/assets/app.json" {
		t.Fatalf("path = %q, want /assets/app.json", item.Path())
	}
	if item.Size() != int64(len(data)) {
		t.Fatalf("size = %d, want %d", item.Size(), len(data))
	}
	if item.ContentType() != "application/json" {
		t.Fatalf("content type = %q, want application/json", item.ContentType())
	}
	if item.ETag() == "" {
		t.Fatal("etag should not be empty")
	}
}

func TestFSProviderRejectsUnsafeAndDirectoryPaths(t *testing.T) {
	t.Parallel()

	provider, err := NewFSProvider(fstest.MapFS{
		"docs":       &fstest.MapFile{Mode: 0o755 | fs.ModeDir},
		"docs/a.txt": &fstest.MapFile{Data: []byte("a")},
	})
	if err != nil {
		t.Fatalf("NewFSProvider failed: %v", err)
	}
	if _, err := provider.Open(context.Background(), "/../secret.txt"); !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("traversal err = %v, want ErrInvalidPath", err)
	}
	if _, err := provider.Open(context.Background(), "/docs"); !errors.Is(err, ErrDirectory) {
		t.Fatalf("directory err = %v, want ErrDirectory", err)
	}
	if _, err := provider.Open(context.Background(), "/missing.txt"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing err = %v, want ErrNotFound", err)
	}
}
