package container

import (
	"context"
	"reflect"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/registration"
)

func TestDeploymentInitializesRegisteredServletsOnceInStartupOrder(t *testing.T) {
	t.Parallel()

	var calls []string
	registry := registration.NewRegistry()
	late, err := registry.AddServlet("late", &configRecordingServlet{calls: &calls})
	if err != nil {
		t.Fatalf("AddServlet late failed: %v", err)
	}
	if err := late.SetLoadOnStartup(20); err != nil {
		t.Fatalf("SetLoadOnStartup late failed: %v", err)
	}
	if conflicts, err := late.AddMapping("/late", "/late/*"); err != nil ||
		len(conflicts) != 0 {
		t.Fatalf(
			"AddMapping late conflicts/err = %#v/%v, want none/nil",
			conflicts,
			err,
		)
	}
	early, err := registry.AddServlet("early", &configRecordingServlet{calls: &calls})
	if err != nil {
		t.Fatalf("AddServlet early failed: %v", err)
	}
	if err := early.SetLoadOnStartup(1); err != nil {
		t.Fatalf("SetLoadOnStartup early failed: %v", err)
	}
	if conflicts, err := early.AddMapping("/early"); err != nil || len(conflicts) != 0 {
		t.Fatalf(
			"AddMapping early conflicts/err = %#v/%v, want none/nil",
			conflicts,
			err,
		)
	}
	lazy, err := registry.AddServlet("lazy", &configRecordingServlet{calls: &calls})
	if err != nil {
		t.Fatalf("AddServlet lazy failed: %v", err)
	}
	if conflicts, err := lazy.AddMapping("/lazy"); err != nil || len(conflicts) != 0 {
		t.Fatalf(
			"AddMapping lazy conflicts/err = %#v/%v, want none/nil",
			conflicts,
			err,
		)
	}

	app, err := servlet.NewWebApp("orders")
	if err != nil {
		t.Fatalf("NewWebApp failed: %v", err)
	}
	snapshot, err := registry.Freeze()
	if err != nil {
		t.Fatalf("Freeze failed: %v", err)
	}
	deployment, err := DeploymentFromRegistration(app, snapshot)
	if err != nil {
		t.Fatalf("DeploymentFromRegistration failed: %v", err)
	}
	application, err := NewApplication(context.Background(), deployment)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := application.Stop(context.Background()); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	want := []string{
		"servlet-init:early:",
		"servlet-init:late:",
		"servlet-init:lazy:",
		"servlet-destroy",
		"servlet-destroy",
		"servlet-destroy",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

type configRecordingServlet struct {
	calls *[]string
}

func (s *configRecordingServlet) Init(
	_ context.Context,
	cfg servlet.ServletConfig,
) error {
	value, _ := cfg.InitParam("encoding")
	*s.calls = append(*s.calls, "servlet-init:"+cfg.Name()+":"+value)
	return nil
}

func (s *configRecordingServlet) Serve(
	context.Context,
	*servlet.Request,
	servlet.Response,
) error {
	*s.calls = append(*s.calls, "servlet")
	return nil
}

func (s *configRecordingServlet) Destroy(context.Context) error {
	*s.calls = append(*s.calls, "servlet-destroy")
	return nil
}

type configRecordingFilter struct {
	calls *[]string
}

func (f *configRecordingFilter) Init(
	_ context.Context,
	cfg servlet.FilterConfig,
) error {
	value, _ := cfg.InitParam("level")
	*f.calls = append(*f.calls, "filter-init:"+cfg.Name()+":"+value)
	return nil
}

func (f *configRecordingFilter) Filter(
	ctx context.Context,
	req *servlet.Request,
	res servlet.Response,
	chain servlet.Chain,
) error {
	*f.calls = append(*f.calls, "filter")
	return chain.Next(ctx, req, res)
}

func (f *configRecordingFilter) Destroy(context.Context) error {
	*f.calls = append(*f.calls, "filter-destroy")
	return nil
}

func recordContainerFilter(name string, calls *[]string) servlet.Filter {
	return servlet.FilterFunc(
		func(ctx context.Context, req *servlet.Request, res servlet.Response, chain servlet.Chain) error {
			*calls = append(*calls, name)
			return chain.Next(ctx, req, res)
		},
	)
}

func noopContainerFilter(
	ctx context.Context,
	req *servlet.Request,
	res servlet.Response,
	chain servlet.Chain,
) error {
	return chain.Next(ctx, req, res)
}
