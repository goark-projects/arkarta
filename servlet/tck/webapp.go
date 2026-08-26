package tck

import (
	"context"
	"io"
	"testing"
	"testing/fstest"
	"time"

	"goark.dev/arkarta/servlet"
)

// RunWebAppContext 执行 WebApp/ServletContext 标准能力兼容性测试。
func RunWebAppContext(t *testing.T) {
	t.Helper()
	app, err := servlet.NewWebApp("tck",
		servlet.WithContextPath("/tck"),
		servlet.WithVirtualServerName("tck.local"),
		servlet.WithSessionTimeout(20*time.Minute),
		servlet.WithMimeType("tck", "application/x-tck"),
		servlet.WithResourceFS(fstest.MapFS{"assets/tck.txt": &fstest.MapFile{Data: []byte("ok")}}),
	)
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	if app.ContextPath() != "/tck" {
		t.Fatalf("context path = %q, want /tck", app.ContextPath())
	}
	if app.VirtualServerName() != "tck.local" {
		t.Fatalf("virtual server = %q, want tck.local", app.VirtualServerName())
	}
	if app.SessionTimeout() != 20*time.Minute {
		t.Fatalf("session timeout = %s, want 20m", app.SessionTimeout())
	}
	if app.MimeType("probe.tck") != "application/x-tck" {
		t.Fatalf("mime type = %q, want application/x-tck", app.MimeType("probe.tck"))
	}
	if app.EffectiveMajorVersion() != servlet.ServletSpecMajorVersion || app.EffectiveMinorVersion() != servlet.ServletSpecMinorVersion {
		t.Fatalf("effective version = %d.%d, want %d.%d",
			app.EffectiveMajorVersion(),
			app.EffectiveMinorVersion(),
			servlet.ServletSpecMajorVersion,
			servlet.ServletSpecMinorVersion,
		)
	}
	file, err := app.OpenResource(context.Background(), "/assets/tck.txt")
	if err != nil {
		t.Fatalf("OpenResource failed: %v", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil || string(data) != "ok" {
		t.Fatalf("resource data = %q/%v, want ok/nil", string(data), err)
	}
}
