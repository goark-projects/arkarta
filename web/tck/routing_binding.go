package tck

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/web"
)

// RunRoutingBinding 执行 Web 路由分组、自动方法和表单绑定兼容性测试。
func RunRoutingBinding(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	t.Run("routes_grouped_handlers", func(t *testing.T) {
		runGroupedRoute(t, factory)
	})
	t.Run("handles_head_options_and_allow", func(t *testing.T) {
		runAutomaticMethods(t, factory)
	})
	t.Run("binds_form_and_converts_parameters", func(t *testing.T) {
		runFormBinding(t, factory)
	})
}

func runGroupedRoute(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	router := web.NewRouter()
	api := router.Group("/api")
	api.Use(web.InterceptorFunc(func(ctx *web.Context, next web.Handler) (web.Result, error) {
		ctx.Response().Header().Set("X-Group", "api")
		return next.Handle(ctx)
	}))
	if err := api.Handle(http.MethodGet, "/users/{id}", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		return web.Text(http.StatusOK, ctx.PathValue("id")), nil
	})); err != nil {
		t.Fatalf("Handle(GET /users/{id}) failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	factory(router).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/users/9", nil))

	if recorder.Code != http.StatusOK || recorder.Body.String() != "9" {
		t.Fatalf("response = %d %q, want 200 9", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("X-Group") != "api" {
		t.Fatalf("X-Group = %q, want api", recorder.Header().Get("X-Group"))
	}
}

func runAutomaticMethods(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	router := web.NewRouter()
	mustHandle(t, router, http.MethodGet, "/health", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		return web.Text(http.StatusOK, "UP"), nil
	}))

	headRecorder := httptest.NewRecorder()
	factory(router).ServeHTTP(headRecorder, httptest.NewRequest(http.MethodHead, "/health", nil))
	if headRecorder.Code != http.StatusOK || headRecorder.Body.Len() != 0 {
		t.Fatalf("HEAD response = %d %q, want 200 empty", headRecorder.Code, headRecorder.Body.String())
	}

	optionsRecorder := httptest.NewRecorder()
	factory(router).ServeHTTP(optionsRecorder, httptest.NewRequest(http.MethodOptions, "/health", nil))
	if optionsRecorder.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS status = %d, want 204", optionsRecorder.Code)
	}
	if optionsRecorder.Header().Get("Allow") != "GET, HEAD, OPTIONS" {
		t.Fatalf("Allow = %q, want GET, HEAD, OPTIONS", optionsRecorder.Header().Get("Allow"))
	}
}

func runFormBinding(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	type input struct {
		Name  string `form:"name"`
		Count int    `form:"count"`
	}
	router := web.NewRouter()
	mustHandle(t, router, http.MethodPost, "/items/{id}", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		id, ok, err := ctx.PathInt("id")
		if err != nil || !ok {
			return nil, err
		}
		var body input
		if err := ctx.BindForm(&body); err != nil {
			return nil, err
		}
		return web.JSON(http.StatusOK, map[string]any{
			"id":    id,
			"name":  body.Name,
			"count": body.Count,
		}), nil
	}))

	values := url.Values{}
	values.Set("name", "arkarta")
	values.Set("count", "3")
	request := httptest.NewRequest(http.MethodPost, "/items/42", strings.NewReader(values.Encode()))
	request.Header.Set("Accept", arkjson.ContentType)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	factory(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	if err := arkjson.Unmarshal(nil, recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response json invalid: %v", err)
	}
	if payload.ID != 42 || payload.Name != "arkarta" || payload.Count != 3 {
		t.Fatalf("payload = %#v, want converted form", payload)
	}
}
