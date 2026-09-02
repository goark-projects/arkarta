package servlet

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestDispatcherForwardUsesTargetPathAndRestoresRequest(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/target", HandlerFunc(func(_ context.Context, req *Request, res Response) error {
		if req.DispatchType() != DispatchForward {
			t.Fatalf("dispatch = %v, want forward", req.DispatchType())
		}
		value, ok := req.Attribute(AttributeForwardRequestURI)
		if !ok || value != "/source" {
			t.Fatalf("forward request uri = %v/%v, want /source/true", value, ok)
		}
		query, _ := req.Attribute(AttributeForwardQueryString)
		if query != "from=source" {
			t.Fatalf("forward query = %v, want from=source", query)
		}
		_, err := res.WriteString(req.Path())
		return err
	}))

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/source?from=source", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	response := newTestResponse()
	dispatcher, err := NewRequestDispatcher(router, "/target")
	if err != nil {
		t.Fatalf("NewRequestDispatcher failed: %v", err)
	}

	if err := dispatcher.Forward(context.Background(), req, response); err != nil {
		t.Fatalf("Forward failed: %v", err)
	}
	if response.body.String() != "/target" {
		t.Fatalf("body = %q, want /target", response.body.String())
	}
	if req.Path() != "/source" {
		t.Fatalf("path = %s, want /source", req.Path())
	}
	if req.DispatchType() != DispatchRequest {
		t.Fatalf("dispatch = %v, want request", req.DispatchType())
	}
}

func TestDispatcherForwardRejectsCommittedResponse(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/target", HandlerFunc(func(context.Context, *Request, Response) error {
		return nil
	}))
	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/source", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	response := newTestResponse()
	if _, err := response.WriteString("committed"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	dispatcher, err := NewRequestDispatcher(router, "/target")
	if err != nil {
		t.Fatalf("NewRequestDispatcher failed: %v", err)
	}

	err = dispatcher.Forward(context.Background(), req, response)
	if !errors.Is(err, ErrResponseCommitted) {
		t.Fatalf("Forward err = %v, want ErrResponseCommitted", err)
	}
}

func TestDispatcherForwardAppliesQueryString(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/target", HandlerFunc(func(_ context.Context, req *Request, _ Response) error {
		if req.QueryString() != "from=dispatcher&x=1" {
			t.Fatalf("query string = %q, want dispatcher query", req.QueryString())
		}
		if req.Query().Get("from") != "dispatcher" {
			t.Fatalf("from = %q, want dispatcher", req.Query().Get("from"))
		}
		value, ok, err := req.Parameter("x")
		if err != nil {
			t.Fatalf("Parameter failed: %v", err)
		}
		if !ok || value != "1" {
			t.Fatalf("x = %q/%v, want 1/true", value, ok)
		}
		return nil
	}))

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/source?from=source", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	dispatcher, err := NewRequestDispatcher(router, "/target?from=dispatcher&x=1")
	if err != nil {
		t.Fatalf("NewRequestDispatcher failed: %v", err)
	}

	if err := dispatcher.Forward(context.Background(), req, newTestResponse()); err != nil {
		t.Fatalf("Forward failed: %v", err)
	}
	if req.QueryString() != "from=source" {
		t.Fatalf("query string restored = %q, want from=source", req.QueryString())
	}
}

func TestDispatcherIncludeCannotChangeOuterStatusOrHeaders(t *testing.T) {
	t.Parallel()

	router := NewRouter()
	mustHandle(t, router, "/fragment", HandlerFunc(func(_ context.Context, req *Request, res Response) error {
		if req.DispatchType() != DispatchInclude {
			t.Fatalf("dispatch = %v, want include", req.DispatchType())
		}
		value, ok := req.Attribute(AttributeIncludeRequestURI)
		if !ok || value != "/page" {
			t.Fatalf("include request uri = %v/%v, want /page/true", value, ok)
		}
		query, _ := req.Attribute(AttributeIncludeQueryString)
		if query != "mode=full" {
			t.Fatalf("include query = %v, want mode=full", query)
		}
		res.SetStatus(http.StatusCreated)
		res.Header().Set("X-Include", "ignored")
		_, err := res.WriteString("fragment")
		return err
	}))

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/page?mode=full", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	response := newTestResponse()
	response.SetStatus(http.StatusAccepted)
	dispatcher, err := NewRequestDispatcher(router, "/fragment")
	if err != nil {
		t.Fatalf("NewRequestDispatcher failed: %v", err)
	}

	if err := dispatcher.Include(context.Background(), req, response); err != nil {
		t.Fatalf("Include failed: %v", err)
	}
	if response.Status() != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Status(), http.StatusAccepted)
	}
	if response.Header().Get("X-Include") != "" {
		t.Fatalf("include header leaked: %q", response.Header().Get("X-Include"))
	}
	if response.body.String() != "fragment" {
		t.Fatalf("body = %q, want fragment", response.body.String())
	}
}

func TestDispatcherErrorSetsErrorAttributes(t *testing.T) {
	t.Parallel()

	cause := errors.New("boom")
	router := NewRouter()
	mustHandle(t, router, "/error", HandlerFunc(func(_ context.Context, req *Request, res Response) error {
		status, _ := req.Attribute(AttributeErrorStatusCode)
		errValue, _ := req.Attribute(AttributeErrorException)
		path, _ := req.Attribute(AttributeErrorRequestURI)
		query, _ := req.Attribute(AttributeErrorQueryString)
		message, _ := req.Attribute(AttributeErrorMessage)
		exceptionType, _ := req.Attribute(AttributeErrorExceptionType)
		if status != http.StatusBadGateway || errValue != cause || path != "/upstream" {
			t.Fatalf("error attrs = %v/%v/%v", status, errValue, path)
		}
		if query != "trace=1" || message != http.StatusText(http.StatusBadGateway) || exceptionType != reflect.TypeOf(cause).String() {
			t.Fatalf("error extended attrs = %v/%v/%v", query, message, exceptionType)
		}
		_, err := res.WriteString("error")
		return err
	}))

	req, err := NewRequest(httptest.NewRequest(http.MethodGet, "/upstream?trace=1", nil))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	response := newTestResponse()
	dispatcher, err := NewRequestDispatcher(router, "/error")
	if err != nil {
		t.Fatalf("NewRequestDispatcher failed: %v", err)
	}

	if err := dispatcher.Error(context.Background(), req, response, http.StatusBadGateway, cause); err != nil {
		t.Fatalf("Error failed: %v", err)
	}
	if response.Status() != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", response.Status(), http.StatusBadGateway)
	}
	if response.body.String() != "error" {
		t.Fatalf("body = %q, want error", response.body.String())
	}
}

type testResponse struct {
	header    Header
	status    int
	committed bool
	body      bytes.Buffer
}

func newTestResponse() *testResponse {
	return &testResponse{
		header: NewHeader(),
		status: http.StatusOK,
	}
}

func (r *testResponse) Header() Header {
	return r.header
}

func (r *testResponse) SetStatus(code int) {
	if !r.committed {
		r.status = code
	}
}

func (r *testResponse) Status() int {
	return r.status
}

func (r *testResponse) Write(data []byte) (int, error) {
	r.committed = true
	return r.body.Write(data)
}

func (r *testResponse) WriteString(value string) (int, error) {
	return r.Write([]byte(value))
}

func (r *testResponse) Flush() error {
	r.committed = true
	return nil
}

func (r *testResponse) Committed() bool {
	return r.committed
}

func (r *testResponse) Reset() error {
	if r.committed {
		return ErrResponseCommitted
	}
	r.header = NewHeader()
	r.status = http.StatusOK
	r.body.Reset()
	return nil
}

func (r *testResponse) BodyWriter() io.Writer {
	return r
}
