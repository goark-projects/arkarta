package web_test

import (
	"encoding/json"
	"io"
	stdmultipart "mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	arkjson "goark.dev/arkarta/json"
	servletmultipart "goark.dev/arkarta/servlet/multipart"
	"goark.dev/arkarta/servlet/nethttp"
	"goark.dev/arkarta/web"
)

func TestRouterGroupUsesScopedInterceptorsAndConvenienceMethods(t *testing.T) {
	t.Parallel()

	var order []string
	router := web.NewRouter()
	router.Use(web.InterceptorFunc(func(ctx *web.Context, next web.Handler) (web.Result, error) {
		order = append(order, "global-before")
		result, err := next.Handle(ctx)
		order = append(order, "global-after")
		return result, err
	}))

	api := router.Group("/api")
	api.Use(web.InterceptorFunc(func(ctx *web.Context, next web.Handler) (web.Result, error) {
		order = append(order, "group-before")
		result, err := next.Handle(ctx)
		order = append(order, "group-after")
		return result, err
	}))
	if err := api.GET("/users/{id}", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		order = append(order, "handler")
		return web.Text(http.StatusOK, ctx.PathValue("id")), nil
	})); err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	nethttp.Handler(router).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/users/42", nil))

	if recorder.Code != http.StatusOK || recorder.Body.String() != "42" {
		t.Fatalf("response = %d %q, want 200 42", recorder.Code, recorder.Body.String())
	}
	wantOrder := []string{"global-before", "group-before", "handler", "group-after", "global-after"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("order = %#v, want %#v", order, wantOrder)
	}
}

func TestRouterProvidesAutomaticHeadOptionsAndAllow(t *testing.T) {
	t.Parallel()

	router := web.NewRouter()
	if err := router.GET("/health", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		return web.Text(http.StatusOK, "UP"), nil
	})); err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	headRecorder := httptest.NewRecorder()
	nethttp.Handler(router).ServeHTTP(headRecorder, httptest.NewRequest(http.MethodHead, "/health", nil))
	if headRecorder.Code != http.StatusOK {
		t.Fatalf("HEAD status = %d, want 200", headRecorder.Code)
	}
	if headRecorder.Body.Len() != 0 {
		t.Fatalf("HEAD body = %q, want empty", headRecorder.Body.String())
	}

	optionsRecorder := httptest.NewRecorder()
	nethttp.Handler(router).ServeHTTP(optionsRecorder, httptest.NewRequest(http.MethodOptions, "/health", nil))
	if optionsRecorder.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS status = %d, want 204", optionsRecorder.Code)
	}
	if got := optionsRecorder.Header().Get("Allow"); got != "GET, HEAD, OPTIONS" {
		t.Fatalf("OPTIONS Allow = %q, want GET, HEAD, OPTIONS", got)
	}

	postRecorder := httptest.NewRecorder()
	postRequest := httptest.NewRequest(http.MethodPost, "/health", nil)
	postRequest.Header.Set("Accept", arkjson.ContentType)
	nethttp.Handler(router).ServeHTTP(postRecorder, postRequest)
	if postRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want 405", postRecorder.Code)
	}
	if got := postRecorder.Header().Get("Allow"); got != "GET, HEAD, OPTIONS" {
		t.Fatalf("POST Allow = %q, want GET, HEAD, OPTIONS", got)
	}
}

func TestContextBindsFormAndConvertsParameters(t *testing.T) {
	t.Parallel()

	type formInput struct {
		Name  string   `form:"name"`
		Count int      `form:"count"`
		Tags  []string `form:"tag"`
	}
	router := web.NewRouter()
	if err := router.POST("/items/{id}", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		id, ok, err := ctx.PathInt("id")
		if err != nil || !ok {
			return nil, err
		}
		trace, ok, err := ctx.QueryBool("trace")
		if err != nil || !ok {
			return nil, err
		}
		var input formInput
		if err := ctx.BindForm(&input); err != nil {
			return nil, err
		}
		return web.JSON(http.StatusOK, map[string]any{
			"id":    id,
			"trace": trace,
			"name":  input.Name,
			"count": input.Count,
			"tags":  input.Tags,
		}), nil
	})); err != nil {
		t.Fatalf("POST failed: %v", err)
	}

	form := url.Values{}
	form.Set("name", "arkarta")
	form.Set("count", "3")
	form.Add("tag", "servlet")
	form.Add("tag", "web")
	request := httptest.NewRequest(http.MethodPost, "/items/42?trace=true", strings.NewReader(form.Encode()))
	request.Header.Set("Accept", arkjson.ContentType)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	nethttp.Handler(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		ID    int      `json:"id"`
		Trace bool     `json:"trace"`
		Name  string   `json:"name"`
		Count int      `json:"count"`
		Tags  []string `json:"tags"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response json invalid: %v", err)
	}
	if payload.ID != 42 || !payload.Trace || payload.Name != "arkarta" || payload.Count != 3 || !reflect.DeepEqual(payload.Tags, []string{"servlet", "web"}) {
		t.Fatalf("payload = %#v, want converted form values", payload)
	}
}

func TestContextBindsMultipartValuesAndParts(t *testing.T) {
	t.Parallel()

	type uploadInput struct {
		Title  string                `form:"title"`
		Avatar servletmultipart.Part `multipart:"avatar"`
	}
	router := web.NewRouter()
	if err := router.POST("/uploads", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		var input uploadInput
		if err := ctx.BindMultipart(&input); err != nil {
			return nil, err
		}
		file, err := input.Avatar.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		return web.JSON(http.StatusCreated, map[string]any{
			"title":    input.Title,
			"filename": input.Avatar.SubmittedFileName(),
			"size":     input.Avatar.Size(),
			"body":     string(data),
		}), nil
	})); err != nil {
		t.Fatalf("POST failed: %v", err)
	}

	var body strings.Builder
	writer := stdmultipart.NewWriter(&body)
	if err := writer.WriteField("title", "avatar"); err != nil {
		t.Fatalf("WriteField failed: %v", err)
	}
	part, err := writer.CreateFormFile("avatar", "profile.txt")
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write([]byte("hello")); err != nil {
		t.Fatalf("part write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart close failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/uploads", strings.NewReader(body.String()))
	request.Header.Set("Accept", arkjson.ContentType)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	nethttp.Handler(router).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", recorder.Code, recorder.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response json invalid: %v", err)
	}
	if payload["title"] != "avatar" || payload["filename"] != "profile.txt" || payload["body"] != "hello" || payload["size"].(float64) != 5 {
		t.Fatalf("payload = %#v, want multipart value and file", payload)
	}
}

func TestRouterAppliesResponseAdviceBeforeWrite(t *testing.T) {
	t.Parallel()

	router := web.NewRouter(web.WithResponseAdvice(web.ResponseAdviceFunc(func(ctx *web.Context, result web.Result) (web.Result, error) {
		ctx.Response().Header().Set("X-Advised", "true")
		return web.Text(http.StatusAccepted, "advised"), nil
	})))
	if err := router.GET("/advice", web.HandlerFunc(func(ctx *web.Context) (web.Result, error) {
		return web.Text(http.StatusOK, "origin"), nil
	})); err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	recorder := httptest.NewRecorder()
	nethttp.Handler(router).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/advice", nil))

	if recorder.Code != http.StatusAccepted || recorder.Body.String() != "advised" {
		t.Fatalf("response = %d %q, want 202 advised", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("X-Advised"); got != "true" {
		t.Fatalf("X-Advised = %q, want true", got)
	}
}
