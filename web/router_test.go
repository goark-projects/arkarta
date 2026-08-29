package web_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	arkjson "goark.dev/arkarta/json"
	"goark.dev/arkarta/servlet/nethttp"
	"goark.dev/arkarta/validation"
	"goark.dev/arkarta/web"
)

type defaultOnlyValidator struct{}

func (defaultOnlyValidator) Validate(context.Context, any) (validation.Result, error) {
	return validation.NewResult(), nil
}

func TestRouterBindsJSONValidatesAndWritesJSON(t *testing.T) {
	t.Parallel()

	type createUserRequest struct {
		Name string `json:"name" arkarta:"required,min=2"`
	}
	router := web.NewRouter(
		web.WithJSONCodec(arkjson.NewCodec(arkjson.WithDisallowUnknownFields(true))),
		web.WithValidator(validation.NewValidator()),
	)
	if err := router.Handle(http.MethodPost, "/users/{id}", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		if got := ctx.PathValue("id"); got != "42" {
			t.Fatalf("path id = %q, want 42", got)
		}
		if got := ctx.QueryValue("trace"); got != "on" {
			t.Fatalf("query trace = %q, want on", got)
		}
		var input createUserRequest
		if err := ctx.BindAndValidateJSON(&input); err != nil {
			return nil, err
		}
		return web.JSON(http.StatusCreated, map[string]string{
			"id":   ctx.PathValue("id"),
			"name": input.Name,
		}), nil
	})); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/users/42?trace=on", strings.NewReader(`{"name":"arkarta"}`))
	request.Header.Set("Accept", arkjson.ContentType)
	request.Header.Set("Content-Type", arkjson.ContentType)
	nethttp.Handler(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != arkjson.ContentType {
		t.Fatalf("content type = %q, want application/json", got)
	}
	var payload map[string]string
	if err := arkjson.Unmarshal(nil, recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response json invalid: %v", err)
	}
	want := map[string]string{"id": "42", "name": "arkarta"}
	if !reflect.DeepEqual(payload, want) {
		t.Fatalf("payload = %#v, want %#v", payload, want)
	}
}

func TestRouterBindAndValidateJSONGroups(t *testing.T) {
	t.Parallel()

	type createUserRequest struct {
		Name string `json:"name" arkarta:"required" arkarta-groups:"create"`
		Code string `json:"code" arkarta:"required"`
	}
	router := web.NewRouter(web.WithValidator(validation.NewValidator()))
	if err := router.Handle(http.MethodPost, "/users", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		var input createUserRequest
		if err := ctx.BindAndValidateJSONGroups(&input, "create"); err != nil {
			return nil, err
		}
		return web.JSON(http.StatusCreated, map[string]string{"name": input.Name}), nil
	})); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	successRecorder := httptest.NewRecorder()
	successRequest := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"arkarta"}`))
	successRequest.Header.Set("Content-Type", arkjson.ContentType)
	nethttp.Handler(router).ServeHTTP(successRecorder, successRequest)
	if successRecorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", successRecorder.Code, successRecorder.Body.String())
	}

	failedRecorder := httptest.NewRecorder()
	failedRequest := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{}`))
	failedRequest.Header.Set("Content-Type", arkjson.ContentType)
	nethttp.Handler(router).ServeHTTP(failedRecorder, failedRequest)
	if failedRecorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422, body=%s", failedRecorder.Code, failedRecorder.Body.String())
	}
	var payload struct {
		Error struct {
			Details []struct {
				Path string `json:"path"`
				Code string `json:"code"`
			} `json:"details"`
		} `json:"error"`
	}
	if err := arkjson.Unmarshal(nil, failedRecorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("error response json invalid: %v", err)
	}
	if len(payload.Error.Details) != 1 || payload.Error.Details[0].Path != "name" || payload.Error.Details[0].Code != "required" {
		t.Fatalf("details = %#v, want only create-group name violation", payload.Error.Details)
	}
}

func TestContextValidateGroupsRequiresGroupValidator(t *testing.T) {
	t.Parallel()

	router := web.NewRouter(web.WithValidator(defaultOnlyValidator{}))
	if err := router.Handle(http.MethodGet, "/validate", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		var input struct{}
		if _, err := ctx.ValidateGroups(&input, "create"); !errors.Is(err, validation.ErrUnsupportedGroups) {
			t.Fatalf("ValidateGroups err = %v, want ErrUnsupportedGroups", err)
		}
		return web.NoContent(), nil
	})); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	nethttp.Handler(router).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/validate", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204, body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestRouterMapsValidationErrors(t *testing.T) {
	t.Parallel()

	type createUserRequest struct {
		Name string `json:"name" arkarta:"required,min=2"`
	}
	router := web.NewRouter(web.WithValidator(validation.NewValidator()))
	if err := router.Handle(http.MethodPost, "/users", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		var input createUserRequest
		if err := ctx.BindAndValidateJSON(&input); err != nil {
			return nil, err
		}
		return web.NoContent(), nil
	})); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":""}`))
	request.Header.Set("Accept", arkjson.ContentType)
	request.Header.Set("Content-Type", arkjson.ContentType)
	nethttp.Handler(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422, body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Details []struct {
				Path string `json:"path"`
				Code string `json:"code"`
			} `json:"details"`
		} `json:"error"`
	}
	if err := arkjson.Unmarshal(nil, recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("error response json invalid: %v", err)
	}
	if payload.Error.Code != "VALIDATION_ERROR" || len(payload.Error.Details) == 0 {
		t.Fatalf("error payload = %#v", payload.Error)
	}
	found := false
	for _, detail := range payload.Error.Details {
		if detail.Path == "name" && detail.Code == "required" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("violations = %#v, want name required", payload.Error.Details)
	}
}

func TestRouterReturnsMethodNotAllowed(t *testing.T) {
	t.Parallel()

	router := web.NewRouter()
	if err := router.Handle(http.MethodGet, "/users/{id}", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		return web.NoContent(), nil
	})); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/users/42", nil)
	request.Header.Set("Accept", arkjson.ContentType)
	nethttp.Handler(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", recorder.Code)
	}
	if got := recorder.Header().Get("Allow"); got != "GET, HEAD, OPTIONS" {
		t.Fatalf("allow = %q, want GET, HEAD, OPTIONS", got)
	}
}

func TestRouterRunsInterceptorsAroundHandler(t *testing.T) {
	t.Parallel()

	var order []string
	router := web.NewRouter()
	router.Use(web.InterceptorFunc(func(ctx *web.Context, next web.Handler) (web.Result, error) {
		order = append(order, "before")
		result, err := next.Handle(ctx)
		order = append(order, "after")
		ctx.Response().Header().Set("X-Trace", "ok")
		return result, err
	}))
	if err := router.Handle(http.MethodGet, "/health", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		order = append(order, "handler")
		return web.Text(http.StatusOK, "UP"), nil
	})); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	nethttp.Handler(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "UP" {
		t.Fatalf("response = %d %q, want 200 UP", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("X-Trace"); got != "ok" {
		t.Fatalf("X-Trace = %q, want ok", got)
	}
	if !reflect.DeepEqual(order, []string{"before", "handler", "after"}) {
		t.Fatalf("order = %#v", order)
	}
}
