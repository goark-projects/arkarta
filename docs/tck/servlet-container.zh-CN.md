# Servlet 容器 TCK 接入指南

语言：[English](servlet-container.md) | 简体中文

本文档面向 Arkarta 容器实现者。容器只应声明自己已经通过 TCK 的 Profile。

## 1. 规则

- Core Profile 是每个 Servlet 容器必须支持的最小 Profile。
- Session、Multipart、Async、Upgrade、Security、Native I/O、Web、WebSocket、JSON 等 Profile 独立声明。
- TCK 测试应在容器仓库中运行，而不是放在 Arkarta 标准仓库内。
- 测试必须使用真实容器适配器或部署入口，不得绕过生命周期直接调用用户 handler。

## 2. Core HTTP

如果容器可以把 `servlet.Handler` 暴露为 `http.Handler`，运行：

```go
func TestCoreHTTP(t *testing.T) {
	tck.RunCoreHTTP(t, func(handler servlet.Handler) http.Handler {
		return mycontainer.NewHandler(handler)
	})
}
```

该测试覆盖状态码、Header、Cookie、Redirect、Error、panic 恢复、Filter、路径映射、请求参数、参数名、Cookie 和请求映射细节。

## 3. 生命周期与部署

实现 `servlet/container.Application` 的容器应运行：

```go
func TestLifecycle(t *testing.T) {
	tck.RunLifecycle(t, func(deployment *container.Deployment) (container.Application, error) {
		return mycontainer.NewApplication(t.Context(), deployment)
	})
}
```

容器必须在请求处理前初始化 Servlet 和 Filter，并在已接受请求完成后销毁它们。

## 4. 错误页

如果容器支持 `servlet.ErrorPageRegistry`：

```go
func TestErrorPages(t *testing.T) {
	tck.RunErrorPages(t, func(handler servlet.Handler, registry *servlet.ErrorPageRegistry) http.Handler {
		return mycontainer.NewHandler(handler, mycontainer.WithErrorPages(registry))
	})
}
```

该测试覆盖状态码错误页、panic 错误页、默认错误页、错误类型优先级和循环保护。

## 5. Session Profile

自定义 `session.Manager` 应运行：

```go
func TestSessionManager(t *testing.T) {
	tck.RunSessionManager(t, func() session.Manager {
		return mycontainer.NewSessionManager()
	})
}
```

如果容器使用 Arkarta 内存参考实现：

```go
func TestMemorySessionProfile(t *testing.T) {
	tck.RunMemorySessionProfile(t, func(options ...session.MemoryManagerOption) *session.MemoryManager {
		return session.NewMemoryManager(options...)
	})
}
```

请求绑定和 URL rewriting 通过 `RunSessionRequestBinding` 验证。

## 6. Multipart Profile

```go
func TestMultipart(t *testing.T) {
	tck.RunMultipartParser(t, func(options ...multipart.Option) *multipart.Parser {
		return multipart.NewParser(options...)
	})
}
```

该测试覆盖普通字段、文件、请求体限制、临时目录、提交文件名清理、Part 删除和表单清理。

## 7. Async、Security 与 Native I/O

```go
func TestAsyncLifecycle(t *testing.T) {
	tck.RunAsyncLifecycle(t)
}

func TestSecurity(t *testing.T) {
	tck.RunSecurity(t)
}

func TestNativeIO(t *testing.T) {
	tck.RunNativeIO(t, func() nativeio.Sender {
		return mycontainer.NativeSender()
	})
}
```

Native I/O 可以使用 sendfile、splice、io_uring、kqueue 或平台特定优化，但必须保持可预测的 Go 错误和取消语义。

## 8. 静态资源与 HTTP 容器入口

```go
func TestStaticResources(t *testing.T) {
	tck.RunStaticResources(t, func(handler servlet.Handler) http.Handler {
		return mycontainer.NewHandler(handler)
	})
}

func TestHTTPContainer(t *testing.T) {
	tck.RunHTTPContainer(t, func() tck.HTTPContainer {
		return mycontainer.NewContainer()
	})
}
```

## 9. Web TCK

使用 `goark.dev/arkarta/web/tck` 验证 Servlet 兼容容器上的 Web、JSON、Validation 行为：

```go
func TestWebJSONValidation(t *testing.T) {
	webtck.RunJSONValidation(t, func(handler servlet.Handler) http.Handler {
		return mycontainer.NewHandler(handler)
	})
}

func TestWebRoutingBinding(t *testing.T) {
	webtck.RunRoutingBinding(t, func(handler servlet.Handler) http.Handler {
		return mycontainer.NewHandler(handler)
	})
}
```

这些测试覆盖 JSON 绑定、Validation 映射、内容协商、405 行为、路由分组、自动 HEAD/OPTIONS、Form 绑定和参数转换。

## 10. JSON Codec TCK

替代 JSON 实现应通过 `goark.dev/arkarta/json/tck`：

```go
func TestJSONCodec(t *testing.T) {
	jsontck.RunCodec(t, jsontck.CodecFactory{
		New: func() arkjson.Codec {
			return myjson.NewCodec()
		},
		WithEscapeHTML: func(enabled bool) arkjson.Codec {
			return myjson.NewCodec(myjson.WithEscapeHTML(enabled))
		},
		WithMaxBytes: func(maxBytes int64) arkjson.Codec {
			return myjson.NewCodec(myjson.WithMaxBytes(maxBytes))
		},
		WithUnknownFieldGate: func(enabled bool) arkjson.Codec {
			return myjson.NewCodec(myjson.WithDisallowUnknownFields(enabled))
		},
		WithUseNumber: func(enabled bool) arkjson.Codec {
			return myjson.NewCodec(myjson.WithUseNumber(enabled))
		},
	})
}
```

## 11. WebSocket TCK

WebSocket 实现应运行：

```go
func TestWebSocketHandshake(t *testing.T) {
	wstck.RunHandshake(t, func(options ...websocket.HandshakeOption) *websocket.Handshaker {
		return websocket.NewHandshaker(options...)
	})
}

func TestWebSocketEndpoint(t *testing.T) {
	wstck.RunEndpointLifecycle(t)
}

func TestWebSocketCompression(t *testing.T) {
	wstck.RunCompression(t)
}

func TestWebSocketFrameCodec(t *testing.T) {
	wstck.RunFrameCodec(t)
}
```

Servlet 集成应使用 `goark.dev/arkarta/websocket/servlet` 完成 HTTP 101 写出、Servlet Upgrade 交接、帧连接适配和 Endpoint 服务循环。

## 12. 发布门禁

声明兼容或发布容器版本前运行：

```shell
go test ./...
go test -race ./...
go vet ./...
gofmt -l .
```

`gofmt -l .` 必须无输出。每项兼容性声明都应列出已经通过的 TCK 入口。
