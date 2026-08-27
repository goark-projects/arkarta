# Arkarta Servlet TCK 接入指南

本文档面向 Arkarta Servlet 容器实现者。容器只有通过对应 Profile 的 TCK 后，才应声明兼容 Arkarta Servlet 1.0。

## 1. 基本原则

- Core Profile 是所有容器必须通过的最小集合。
- Session、Multipart、Async、Upgrade、Security、Native I/O 等 Profile 只能在实现对应能力并通过测试后声明。
- TCK 测试应放在容器仓库自己的测试包中运行，避免把容器实现反向依赖到 Arkarta 标准仓库。
- 容器测试必须使用真实容器适配入口，不应绕过容器生命周期直接调用应用代码。

## 2. Core HTTP 适配

如果容器可以把 `servlet.Handler` 暴露为 `http.Handler`，应先接入 Core HTTP 测试：

```go
package mycontainer_test

import (
	"net/http"
	"testing"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/tck"
)

func TestCoreHTTP(t *testing.T) {
	tck.RunCoreHTTP(t, func(handler servlet.Handler) http.Handler {
		return mycontainer.NewHandler(handler)
	})
}
```

该测试覆盖响应状态、Header、Cookie、Redirect、错误映射、panic 恢复、过滤器顺序、路径映射、请求参数和映射元素。

## 3. 生命周期与部署

容器实现 `servlet/container.Application` 后，应接入生命周期测试：

```go
func TestLifecycle(t *testing.T) {
	tck.RunLifecycle(t, func(deployment *container.Deployment) (container.Application, error) {
		return mycontainer.NewApplication(t.Context(), deployment)
	})
}
```

该测试要求 Servlet、Filter、RequestListener 和 WebApp 生命周期顺序稳定。容器必须保证 `Init` 早于请求处理，`Destroy` 晚于已接受请求。

## 4. 错误页

如果容器支持 `servlet.ErrorPageRegistry`，应接入错误页测试：

```go
func TestErrorPages(t *testing.T) {
	tck.RunErrorPages(t, func(handler servlet.Handler, registry *servlet.ErrorPageRegistry) http.Handler {
		return mycontainer.NewHandler(handler, mycontainer.WithErrorPages(registry))
	})
}
```

该测试覆盖状态码错误页、panic 错误页、默认错误页、错误类型后注册优先和错误页循环保护。

## 5. Session Profile

实现 `session.Manager` 后，应接入 Session 管理器测试：

```go
func TestSessionManager(t *testing.T) {
	tck.RunSessionManager(t, func() session.Manager {
		return mycontainer.NewSessionManager()
	})
}
```

如果容器复用或兼容 Arkarta 内存会话参考实现，还应接入：

```go
func TestMemorySessionProfile(t *testing.T) {
	tck.RunMemorySessionProfile(t, func(options ...session.MemoryManagerOption) *session.MemoryManager {
		return session.NewMemoryManager(options...)
	})
}
```

请求绑定能力应通过 `RunSessionRequestBinding` 验证，覆盖创建、加载、ID 轮换和 URL rewriting。

## 6. Multipart Profile

实现 multipart 解析器后，应接入：

```go
func TestMultipart(t *testing.T) {
	tck.RunMultipartParser(t, func(options ...multipart.Option) *multipart.Parser {
		return multipart.NewParser(options...)
	})
}
```

该测试覆盖字段、文件、请求体限制、临时目录、提交文件名归一化、Part 删除和表单清理。

## 7. Async 与 Security

Async Profile 应通过：

```go
func TestAsyncLifecycle(t *testing.T) {
	tck.RunAsyncLifecycle(t)
}
```

Security Profile 应通过：

```go
func TestSecurity(t *testing.T) {
	tck.RunSecurity(t)
}
```

Security 测试覆盖 Basic 认证、Realm、角色映射、方法级约束和 run-as 作用域恢复。

## 8. Native I/O Profile

实现 Native I/O Profile 后，应接入：

```go
func TestNativeIO(t *testing.T) {
	tck.RunNativeIO(t, func() nativeio.Sender {
		return mycontainer.NativeSender()
	})
}
```

该测试覆盖文件区段发送、非法区段拒绝和上下文取消。容器可以用 sendfile、splice、io_uring、kqueue 或系统特定机制优化，但必须保留可预测的 Go 错误语义。

## 9. 静态资源与 HTTP 容器入口

静态资源 default servlet 应通过：

```go
func TestStaticResources(t *testing.T) {
	tck.RunStaticResources(t, func(handler servlet.Handler) http.Handler {
		return mycontainer.NewHandler(handler)
	})
}
```

容器入口应通过：

```go
func TestHTTPContainer(t *testing.T) {
	tck.RunHTTPContainer(t, func() tck.HTTPContainer {
		return mycontainer.NewContainer()
	})
}
```

## 10. 推荐门禁

容器仓库发布前至少运行：

```shell
go test ./...
go test -race ./...
go vet ./...
gofmt -l .
```

`gofmt -l .` 必须无输出。声明兼容某个 Profile 前，必须能指出对应 TCK 入口和测试结果。
