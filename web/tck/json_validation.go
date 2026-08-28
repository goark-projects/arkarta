package tck

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/validation"
	"goark.dev/arkarta/web"
)

// HTTPHandlerFactory 将 Servlet Handler 暴露为标准库 http.Handler。
type HTTPHandlerFactory func(servlet.Handler) http.Handler

// RunJSONValidation 执行 Web、JSON 与 Validation 组合兼容性测试。
func RunJSONValidation(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	t.Run("binds_json_validates_and_writes_json", func(t *testing.T) {
		runBindValidateAndWriteJSON(t, factory)
	})
	t.Run("maps_validation_error", func(t *testing.T) {
		runValidationError(t, factory)
	})
	t.Run("maps_bad_json", func(t *testing.T) {
		runBadJSON(t, factory)
	})
	t.Run("returns_not_acceptable", func(t *testing.T) {
		runNotAcceptable(t, factory)
	})
	t.Run("returns_method_not_allowed", func(t *testing.T) {
		runMethodNotAllowed(t, factory)
	})
}

func runBindValidateAndWriteJSON(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	type input struct {
		Name string `json:"name" arkarta:"required,min=2"`
	}
	router := web.NewRouter(
		web.WithJSONCodec(arkjson.NewStandardCodec(arkjson.WithDisallowUnknownFields(true))),
		web.WithValidator(validation.NewValidator()),
	)
	mustHandle(t, router, http.MethodPost, "/accounts/{id}", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		var body input
		if err := ctx.BindAndValidateJSON(&body); err != nil {
			return nil, err
		}
		return web.JSON(http.StatusCreated, map[string]string{
			"id":   ctx.PathValue("id"),
			"name": body.Name,
		}), nil
	}))

	recorder := httptest.NewRecorder()
	request := jsonRequest(http.MethodPost, "/accounts/7", `{"name":"goark"}`)
	factory(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	payload := decodeBody[map[string]string](t, recorder)
	if payload["id"] != "7" || payload["name"] != "goark" {
		t.Fatalf("payload = %#v, want id/name", payload)
	}
}

func runValidationError(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	type input struct {
		Name string `json:"name" arkarta:"required"`
	}
	router := web.NewRouter()
	mustHandle(t, router, http.MethodPost, "/accounts", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		var body input
		if err := ctx.BindAndValidateJSON(&body); err != nil {
			return nil, err
		}
		return web.NoContent(), nil
	}))

	recorder := httptest.NewRecorder()
	factory(router).ServeHTTP(recorder, jsonRequest(http.MethodPost, "/accounts", `{"name":""}`))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d, body=%s", recorder.Code, http.StatusUnprocessableEntity, recorder.Body.String())
	}
	payload := decodeBody[web.ErrorResponse](t, recorder)
	if payload.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("error code = %q, want VALIDATION_ERROR", payload.Error.Code)
	}
}

func runBadJSON(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	router := web.NewRouter()
	mustHandle(t, router, http.MethodPost, "/accounts", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		var body map[string]string
		if err := ctx.BindJSON(&body); err != nil {
			return nil, err
		}
		return web.NoContent(), nil
	}))

	recorder := httptest.NewRecorder()
	factory(router).ServeHTTP(recorder, jsonRequest(http.MethodPost, "/accounts", `{"name":`))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	payload := decodeBody[web.ErrorResponse](t, recorder)
	if payload.Error.Code != "BAD_REQUEST" {
		t.Fatalf("error code = %q, want BAD_REQUEST", payload.Error.Code)
	}
}

func runNotAcceptable(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	router := web.NewRouter()
	mustHandle(t, router, http.MethodGet, "/accounts", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		return web.JSON(http.StatusOK, map[string]string{"ok": "true"}), nil
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	request.Header.Set("Accept", "application/xml")
	factory(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotAcceptable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotAcceptable)
	}
}

func runMethodNotAllowed(t *testing.T, factory HTTPHandlerFactory) {
	t.Helper()
	router := web.NewRouter()
	mustHandle(t, router, http.MethodGet, "/accounts/{id}", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		return web.NoContent(), nil
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/accounts/7", nil)
	request.Header.Set("Accept", arkjson.ContentType)
	factory(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	if recorder.Header().Get("Allow") != "GET, HEAD, OPTIONS" {
		t.Fatalf("allow = %q, want GET, HEAD, OPTIONS", recorder.Header().Get("Allow"))
	}
}

func jsonRequest(method, target, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Accept", arkjson.ContentType)
	request.Header.Set("Content-Type", arkjson.ContentType)
	return request
}

func mustHandle(t *testing.T, router *web.Router, method, pattern string, handler web.Handler) {
	t.Helper()
	if err := router.Handle(method, pattern, handler); err != nil {
		t.Fatalf("Handle(%s %s) failed: %v", method, pattern, err)
	}
}

func decodeBody[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()
	var payload T
	if err := arkjson.Unmarshal(nil, recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response json invalid: %v, body=%s", err, recorder.Body.String())
	}
	return payload
}
