# Arkarta Servlet 1.0 规范草案

状态：Release Candidate 1
目标模块：`goark.dev/arkarta/servlet`  
规范基线：Jakarta Servlet 6.1 正式规范、Jakarta Servlet 6.2 开发中路线、Java Servlet 4.0 / JSR 369 历史语义
发布日期：2026-08-27

## 1. 目标

Arkarta Servlet 1.0 定义 Goark Web 应用与 Web 容器之间的稳定契约。它不是 Java Servlet API 的逐字翻译，而是吸收 Servlet 体系中已经被 Tomcat、Jetty、WebLogic、GlassFish 等容器验证过的核心语义，并用 Go 的语言模型重新表达。

本标准解决四个问题：

1. 统一 Goark Web 应用的请求、响应、过滤器、生命周期、上下文和错误处理契约。
2. 让不同容器实现同一个标准，例如 Goark Tomcat、Goark Jetty、`net/http` 适配容器和未来的高性能原生容器。
3. 给上层 `goark.dev/arkarta/web`、MVC、Security、WebSocket、静态资源、模板和观测模块提供稳定底座。
4. 通过 TCK 兼容性测试约束容器行为，避免标准只停留在接口声明。

## 2. 规范用语

- 必须：实现或应用必须遵守，违反即不兼容。
- 应当：默认要求，只有存在明确工程理由时才能偏离。
- 可以：可选能力，不影响核心兼容性。
- 容器：实现本标准并负责监听、部署、调度、生命周期和资源管理的运行时。
- 应用：由用户代码组成的 Web 模块，注册处理器、过滤器、监听器和资源。

## 3. 外部标准基线

Arkarta Servlet 1.0 以 Jakarta Servlet 6.1 作为正式基线。Jakarta Servlet 6.2 当前在 Jakarta EE 12 路线下作为维护版演进，Arkarta 1.0 只吸收其中稳定且不破坏 Go API 的方向。Java Servlet 4.0 / JSR 369 作为 Oracle/JCP 体系下的历史基线，主要参考 HTTP/2、异步、升级、TCK 和兼容性声明模式。

Arkarta Servlet 不声明兼容 Jakarta Servlet，也不复用 `jakarta.servlet` 或 `javax.servlet` 命名空间。任何容器只有通过 Arkarta Servlet TCK 后，才能声明兼容 Arkarta Servlet 1.0。

参考资料：

- Jakarta Servlet 6.1: https://jakarta.ee/specifications/servlet/6.1/
- Jakarta Servlet 6.1 Specification: https://jakarta.ee/specifications/servlet/6.1/jakarta-servlet-spec-6.1
- Jakarta Servlet 6.2: https://jakarta.ee/specifications/servlet/6.2/
- Jakarta Servlet 6.2 release record: https://projects.eclipse.org/projects/ee4j.servlet/releases/6.2.0
- JSR 369 Java Servlet 4.0: https://jcp.org/ja/jsr/detail?id=369
- Oracle Java EE 8 technology list: https://www.oracle.com/java/technologies/java-ee-glance.html

## 4. 设计原则

1. Go 优先：公共 API 必须优先使用 Go 标准库、接口组合、显式错误返回和 `context.Context`。
2. 互操作优先：标准层必须能与 `net/http` 无损互通，不能绕开 Go 生态另造完整 HTTP 基础类型。
3. 容器中立：标准不得假定 Tomcat、Jetty、`net/http`、epoll、io_uring 或其他具体实现。
4. 生命周期显式：初始化、启动、停止、销毁和请求取消都必须有明确时序。
5. 零反射依赖：标准不依赖运行时扫描、动态代理或 Java 注解模型。注册可以由代码显式完成，也可以由 Goark CLI 生成。
6. 兼容性可测试：所有必须语义都必须能被 TCK 验证。
7. 小核心，多 Profile：核心标准保持小而稳定，会话、多部分、升级、原生 I/O 等能力通过 Profile 扩展。

## 5. 非目标

Arkarta Servlet 1.0 不包含以下内容：

- JSP、JSTL、Expression Language 或 Java 模板语义。
- Java 注解、`web.xml` 作为强制部署格式、WAR/EAR 打包格式。
- 具体认证提供者、用户目录、OAuth2/JWT、ACL/ABAC 策略源。声明式安全约束由 `servlet/security` 定义，企业级认证授权集成由 `goark-security` 或上层 Web 模块承载。
- Java 式类继承、`GenericServlet`、`HttpServlet` 的继承层级。
- Java IO/NIO 类型、`ByteBuffer` 类型和线程池模型。
- 对 HTTP/2 Server Push 的强制要求。容器可以提供资源提示或协议级 Push 能力，但核心标准不强制。

## 6. Profile 划分

Arkarta Servlet 1.0 定义一个必选核心 Profile 和若干可选 Profile。

### 6.1 Core Profile

容器必须实现：

- 请求和响应模型。
- 同步处理器。
- 过滤器链。
- 应用上下文。
- 生命周期。
- 路径映射与默认映射。
- 错误返回、错误分发和 panic 恢复。
- `net/http` 适配。
- TCK Core 测试。

### 6.2 Session Profile

容器可以实现：

- 会话创建、读取、失效和 ID 轮换。
- Cookie 或自定义传输方式。
- 空闲超时。
- 会话属性变更事件。

### 6.3 Multipart Profile

容器可以实现：

- `multipart/form-data` 解析。
- 单文件、总请求体、内存阈值和临时文件目录限制。
- 上传临时文件清理。

### 6.4 Async/Stream Profile

容器可以实现：

- 显式异步请求延长。
- 流式响应。
- 客户端断开传播。
- 超时、取消和完成事件。

### 6.5 Upgrade Profile

容器可以实现：

- HTTP 协议升级。
- WebSocket 或其他升级协议的接入点。
- 对 HTTP/2、HTTP/3 能力的协议无关暴露。

### 6.6 Native I/O Profile

容器可以实现：

- 零拷贝文件发送。
- `sendfile`、`splice`、`io_uring`、`epoll`、`kqueue` 等平台能力。
- 原生连接统计和背压信号。

## 7. 包结构

第一版建议包结构：

```text
arkarta/servlet                 应用侧核心 API
arkarta/servlet/container       容器实现 SPI
arkarta/servlet/nethttp         标准库 net/http 适配
arkarta/servlet/registration    Servlet、Filter、Listener 动态注册元模型
arkarta/servlet/resource        静态资源、default servlet 与 welcome file
arkarta/servlet/tck             容器兼容性测试工具
arkarta/servlet/session         可选会话 Profile
arkarta/servlet/multipart       可选上传 Profile
arkarta/servlet/async           可选异步与流式响应 Profile
arkarta/servlet/upgrade         可选协议升级 Profile
arkarta/servlet/security        可选声明式安全 Profile
arkarta/servlet/nativeio        可选 Native I/O Profile
arkarta/websocket               WebSocket 独立标准包
arkarta/websocket/servlet       WebSocket 与 Servlet Upgrade 适配
```

`servlet` 根包只放稳定、小型、应用高频使用的接口和类型。容器 SPI、TCK、Profile 不得反向污染根包。

## 8. 核心编程模型

Arkarta Servlet 以 `Handler` 作为应用侧最小处理单元。`Servlet` 是带生命周期的处理单元，适合容器管理。

```go
package servlet

import "context"

// Handler 是应用处理请求的最小契约。
type Handler interface {
	Serve(ctx context.Context, req *Request, res Response) error
}

// HandlerFunc 让普通函数可以直接作为处理器使用。
type HandlerFunc func(ctx context.Context, req *Request, res Response) error

// Serve 执行函数式处理器。
func (f HandlerFunc) Serve(ctx context.Context, req *Request, res Response) error {
	return f(ctx, req, res)
}

// Servlet 是带生命周期的容器托管处理器。
type Servlet interface {
	Handler
	Init(ctx context.Context, cfg ServletConfig) error
	Destroy(ctx context.Context) error
}
```

约束：

- 容器必须为每个请求传入有效 `context.Context`。
- `ctx` 取消后，处理器应当尽快停止阻塞操作并返回。
- `Handler` 可以被并发调用，应用实现必须自行保证共享状态并发安全。
- `Servlet.Init` 对同一实例最多调用一次，且必须先于任何请求处理。
- `Servlet.Destroy` 在容器停止或应用卸载时调用，且必须晚于最后一个已接受请求。

## 9. 请求模型

`Request` 是对 Go HTTP 请求语义的标准化封装。

必选能力：

- HTTP 方法、协议、Scheme、Host、Path、Query。
- Header、Cookie、Content-Length、RemoteAddr。
- RequestURI、RequestURL、QueryString、ContextPath、ServletPath、PathInfo、Mapping。
- Query 与 `application/x-www-form-urlencoded` 表单参数合并视图。
- Body 读取与关闭。
- TLS 与安全传输标记。
- 请求属性。
- 分发类型。
- 原始 `*http.Request` 互操作。

草案形态：

```go
package servlet

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// Request 表示容器传给应用的请求视图。
type Request struct {
	// 字段不直接导出，避免应用绕过标准语义。
}

func (r *Request) Context() context.Context
func (r *Request) Method() string
func (r *Request) Protocol() string
func (r *Request) Scheme() string
func (r *Request) Host() string
func (r *Request) RequestURI() string
func (r *Request) RequestURL() string
func (r *Request) QueryString() string
func (r *Request) ContextPath() string
func (r *Request) Path() string
func (r *Request) Query() url.Values
func (r *Request) Parameter(name string) (string, bool, error)
func (r *Request) ParameterValues(name string) ([]string, bool, error)
func (r *Request) ServletPath() string
func (r *Request) PathInfo() string
func (r *Request) Mapping() RequestMapping
func (r *Request) Header() http.Header
func (r *Request) Cookie(name string) (*http.Cookie, error)
func (r *Request) Body() io.ReadCloser
func (r *Request) Attribute(key string) (any, bool)
func (r *Request) SetAttribute(key string, value any)
func (r *Request) DispatchType() DispatchType
func (r *Request) HTTPRequest() *http.Request
```

约束：

- Header 名称必须按 Go `net/http` 规范进行规范化访问。
- Query 必须只反映 URL 查询串；Parameter 视图才合并查询串与表单体参数。
- Body 只能被消费一次，容器不得默认缓存整个请求体。
- 属性键推荐使用反向域名或包路径前缀，避免跨框架冲突。
- `HTTPRequest()` 必须返回与当前请求等价的标准库请求；容器可以返回只读视图。

## 10. 响应模型

`Response` 是应用写出响应的标准接口。

```go
package servlet

import (
	"io"
	"net/http"
)

// Response 表示容器提供的响应写出能力。
type Response interface {
	Header() http.Header
	SetStatus(code int)
	Status() int
	Write([]byte) (int, error)
	WriteString(value string) (int, error)
	Flush() error
	Committed() bool
	Reset() error
	BodyWriter() io.Writer
}
```

约束：

- 未显式设置状态码时，首次写入必须提交 `200 OK`。
- Header 在响应提交后不得再影响已发送部分。
- `Reset` 只能在响应提交前成功。
- 容器必须支持设置标准 HTTP 状态码；新增状态码优先使用 Go 标准库常量，缺失时允许直接传整数。
- Redirect 必须允许调用者控制状态码，并允许写入可选响应体。
- 字符编码通过 `Content-Type` 的 `charset` 参数表达，不引入 Java 风格重载方法。
- `AddCookie`、`Redirect`、`SendError`、`SetContentType`、`SetCharacterEncoding`、`SetContentLength` 是标准便利 API，容器必须保持与底层 `Response` 提交状态一致。

## 11. 过滤器链

过滤器负责横切处理，例如日志、鉴权、压缩、Trace、限流和 CORS。

```go
package servlet

import "context"

// Filter 是请求进入目标处理器前后的横切处理器。
type Filter interface {
	Filter(ctx context.Context, req *Request, res Response, chain Chain) error
}

// ManagedFilter 是带生命周期的容器托管过滤器。
type ManagedFilter interface {
	Filter
	Init(ctx context.Context, cfg FilterConfig) error
	Destroy(ctx context.Context) error
}

// Chain 表示当前过滤器之后的剩余链路。
type Chain interface {
	Next(ctx context.Context, req *Request, res Response) error
}
```

约束：

- 过滤器执行顺序必须确定。
- Filter 初始化必须先于任何请求过滤，Destroy 必须晚于最后一次过滤调用。
- FilterBinding 必须按 DispatchType 过滤，默认只匹配 `DispatchRequest`。
- 过滤器可以不调用 `chain.Next`，用于短路响应。
- 过滤器调用 `chain.Next` 不得超过一次。
- 如果响应已经提交，后续错误不得再触发新的错误页写出，只能由容器记录或关闭连接。
- 容器必须在 TCK 中验证过滤器顺序、短路和错误传播。

## 12. 生命周期

应用生命周期必须遵循：

```text
Deploy -> Init -> Start -> Accept Requests -> Stop -> Destroy -> Undeploy
```

要求：

- `Init` 阶段不得接受请求。
- `Start` 成功后才能接收请求。
- `Stop` 后不得接收新请求，但应当允许已接收请求在关闭超时内完成。
- `Destroy` 必须释放应用资源。
- 容器关闭必须具备超时和强制终止路径。
- 生命周期错误必须被结构化返回，不得只写日志。

## 13. 应用上下文

`WebApp` 或 `ServletContext` 表示一个部署单元的运行时上下文。Goark 1.0 推荐对外使用 `WebApp`，保留 `ServletContext` 作为语义别名或文档术语，避免与 `context.Context` 混淆。

必选能力：

- 应用名、上下文路径。
- 初始化参数。
- 应用级属性。
- 日志入口。
- 资源查找。
- 处理器、过滤器和监听器注册。
- `servlet/registration` 动态注册元模型。

约束：

- 上下文属性必须并发安全。
- 初始化参数在 `Start` 后必须只读。
- 动态注册默认只允许在 `Init` 阶段执行。
- 动态注册元模型必须覆盖名称、实现类型名、初始化参数、异步标记、Servlet URL 映射、Filter URL/Servlet-name 映射、Listener 顺序和冻结快照。
- 运行期热注册属于容器扩展能力，不属于 Core Profile。

## 14. 路径映射与分发

Core Profile 必须支持四类映射：

1. 精确匹配：`/users/me`
2. 前缀匹配：`/api/*`
3. 扩展匹配：`*.json`
4. 默认匹配：`/`

匹配优先级：

```text
精确匹配 > 最长前缀匹配 > 扩展匹配 > 默认匹配
```

分发类型：

```go
package servlet

// DispatchType 表示请求进入处理链的原因。
type DispatchType uint8

const (
	DispatchRequest DispatchType = iota
	DispatchForward
	DispatchInclude
	DispatchError
	DispatchAsync
)
```

要求：

- 容器必须记录当前分发类型。
- 错误分发必须保留原始请求路径、状态码和错误对象。
- Include 不得修改主响应状态码。
- Forward 只能在响应提交前执行。

## 15. 错误模型

Arkarta Servlet 使用显式 `error` 返回，不使用 Java 异常模型。

```go
package servlet

// StatusError 表示可映射到 HTTP 状态码的处理错误。
type StatusError interface {
	error
	StatusCode() int
	PublicMessage() string
}
```

规则：

- `nil` 表示处理成功。
- 返回 `StatusError` 时，容器必须使用其状态码。
- 返回普通错误时，未提交响应必须映射为 `500 Internal Server Error`。
- panic 必须由容器恢复并映射为 `500`，同时记录堆栈。
- 响应已提交后出现错误，容器不得再重写响应体。
- 错误响应不得默认泄露堆栈、文件路径、环境变量或内部配置。

## 16. 会话 Profile

会话不是 Core Profile 的强制能力。实现 Session Profile 的容器必须提供：

- `GetSession(create bool)` 语义。
- `Session.ID()`。
- `Session.Invalidate()`。
- `Session.RenewID()`。
- `session.Accessor.Get(ctx, req, res, create)` 请求绑定语义。
- `session.Accessor.ChangeID(ctx, req, res)` 登录后会话 ID 轮换语义。
- `RequestedID` 与 `RequestedIDValid`。
- WebApp 级 Session Cookie 配置。
- COOKIE、URL、SSL 三类 tracking policy。
- Store SPI、内存 Store、passivation 和 activation 回调。
- 空闲超时。
- 属性增删改查。
- 属性名快照。
- 并发安全。

安全要求：

- 默认 Cookie 必须 `HttpOnly`。
- TLS 请求下默认 Cookie 必须 `Secure`。
- 默认应当使用 `SameSite=Lax`。
- 默认 Cookie 名称为 `JSESSIONID`，默认 Path 按请求 ContextPath 计算，空 ContextPath 使用 `/`。
- URL rewriting 必须使用路径参数 `jsessionid`，并保留 query string 与 fragment；请求已携带有效会话 Cookie 时应优先保持 URL 不变。
- 登录成功后应当轮换 Session ID。
- 容器不得把敏感会话数据写入客户端 Cookie，除非使用明确的加密和认证机制。

## 17. Multipart Profile

实现 Multipart Profile 的容器必须支持：

- 最大请求体大小。
- 最大单文件大小。
- 内存阈值。
- 临时目录。
- 提交文件名归一化。
- Part 读取、写入和删除。
- 文件句柄关闭。
- 请求结束后清理临时文件。

解析失败必须返回结构化错误，不能让应用收到半解析状态。

## 18. Async/Stream Profile

Go 已经用 goroutine 和 `context.Context` 解决了 Java Servlet 中部分异步需求，因此 Goark 不直接复制 `AsyncContext`。

第一版原则：

- Core Profile 中，处理器返回即代表请求生命周期结束。
- 如果应用需要在返回后继续写响应，必须显式进入 Async/Stream Profile。
- Async 必须有完成、超时、取消和错误通知。
- 客户端断开必须传播到 `context.Context`。
- 容器必须防止响应对象在完成后继续被写入。

## 19. 协议升级 Profile

Upgrade Profile 用于 WebSocket、CONNECT、HTTP/2 扩展或未来 HTTP/3 能力。

要求：

- 升级必须在响应提交前完成。
- 升级后连接所有权必须从 Servlet 响应模型转交给升级处理器。
- 容器必须明确连接关闭责任。
- WebSocket 标准不放在 Servlet Core 中，第一版由 `goark.dev/arkarta/websocket` 包承载，Servlet 集成由 `goark.dev/arkarta/websocket/servlet` 适配包承载。

## 20. 容器 SPI

容器实现通过 `servlet/container` 包接入标准。

```go
package container

import (
	"context"

	"goark.dev/arkarta/servlet"
)

// Container 是 Web 容器必须实现的入口。
type Container interface {
	Deploy(ctx context.Context, app *Deployment) (Application, error)
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// Application 表示一个已部署应用。
type Application interface {
	Handler() servlet.Handler
	Context() *servlet.WebApp
	Stop(ctx context.Context) error
}
```

要求：

- `Deploy` 必须校验映射冲突、过滤器顺序、初始化参数和 Profile 依赖。
- `Start` 必须幂等或返回明确错误。
- `Shutdown` 必须停止接受新连接并等待请求完成。
- 容器必须暴露自身名称、版本、支持的 Profile 和限制参数。

## 21. Native I/O Profile

Native I/O Profile 用于把大文件传输、零拷贝、平台事件循环和背压能力标准化。Arkarta 只定义契约，不要求所有容器使用同一种内核机制。

要求：

- `servlet/nativeio.Sender` 必须按 `FileRegion` 的 offset/count 发送数据。
- 发送前必须检查 `context.Context`；发送过程中应当尽力响应取消。
- 发送结果必须返回字节数和实际策略。
- 跨平台参考实现必须提供 buffered fallback。
- 容器只有在暴露可用 Sender 并通过 Native I/O TCK 后，才能声明 `ProfileNativeIO`。
- Linux 容器可以在该契约下接入 `sendfile`、`splice` 或 `io_uring`；BSD/macOS 容器可以接入 `kqueue` 相关能力；Windows 容器可以接入系统支持的文件传输优化。

## 22. Goark 集成边界

Arkarta Servlet 与现有 Goark 模块的关系：

- `goark` core：提供 DI、环境、生命周期、事件等通用基础设施。
- `goark-boot`：负责启动装配、配置加载和容器选择。
- `goark.dev/arkarta/web`：未来提供 MVC、路由、参数绑定、响应编解码等上层开发体验。
- `goark-security`：负责认证、授权、Principal、Session 固定防护和安全过滤器。
- `goark-log` / observability：负责日志、Trace、Metrics。

`servlet` 不直接依赖具体容器，也不强制依赖 `goark-boot`。如果需要接入 Goark core，应通过小型适配包完成，避免标准层被启动框架反向绑定。

## 23. 兼容性测试 TCK

任何容器声明兼容 Arkarta Servlet 1.0 前，必须通过对应 Profile 的 TCK。

Core TCK 必须覆盖：

- 生命周期调用次数和顺序。
- 并发请求下的处理器安全边界。
- 请求 Header、Query、Cookie、Body 语义。
- 响应状态码、Header、提交、Flush、Reset。
- 过滤器顺序、短路、错误传播。
- Filter 生命周期和 DispatcherType 过滤。
- 路径映射优先级。
- Forward、Include、Error 分发。
- Servlet、Filter、Listener 注册元模型与冻结语义。
- 动态注册关闭、冻结快照转换、load-on-startup 顺序和同名 Servlet 单次初始化。
- WebApp/ServletContext 版本、MIME、资源、日志和会话超时。
- 静态资源、GET/HEAD、条件请求、If-Range、多 Range 和 welcome file。
- Session 请求/响应 Cookie 绑定、requested ID 校验、ID 轮换、SSL tracking、CookieConfig、Store 和 activation/passivation。
- Session URL rewriting 和属性名快照。
- Multipart 临时目录、提交文件名归一化、Part 删除和表单清理。
- Async Await、完成幂等、dispatch 计数、超时事件顺序和 Stream 关闭后写入拒绝。
- Security Basic 认证、角色映射、方法约束和 run-as 作用域。
- Native I/O 文件区段发送、非法区段拒绝和上下文取消。
- 错误页默认映射、错误类型优先级和循环保护。
- `context.Context` 取消传播。
- `net/http` 适配一致性。
- panic 恢复与错误响应。
- `go test -race` 下无数据竞争。

Profile TCK 按 Profile 独立运行。容器只能声明自己通过的 Profile。

## 24. 版本策略

- Arkarta Servlet 1.x 保持源代码兼容，新增能力必须通过可选接口、可选方法包装类型或新包完成。
- 不修改已发布方法签名。
- 不改变已发布错误语义。
- 不把可选 Profile 提升为 Core Profile，除非进入新的主版本。
- 弃用周期至少跨一个次版本。

## 25. 第一阶段交付物

第一阶段交付标准骨架、标准库适配和可复用 Profile 实现，不实现具体 Tomcat/Jetty 容器：

1. `go.mod`：模块路径 `goark.dev/arkarta`。
2. `servlet` 子包：`Handler`、`HandlerFunc`、`Servlet`、`Request`、`Response`、`Filter`、`Chain`、`DispatchType`、`StatusError`。
3. `container` 包：`Container`、`Deployment`、`Application`、Profile 声明。
4. `registration` 包：动态注册元模型和冻结快照。
5. `nethttp` 包：从 `net/http` 到 Arkarta Servlet 的适配。
6. `session` 包：Session Profile 接口、请求绑定和内存会话管理器。
7. `resource` 包：静态资源 Provider、default servlet 和 welcome file 解析。
8. `multipart` 包：Multipart Profile 解析器、Part API 和请求绑定。
9. `async` 包：Async/Stream Profile。
10. `upgrade` 包：协议升级 Profile。
11. `security` 包：声明式安全 Profile。
12. `nativeio` 包：Native I/O Profile。
13. `websocket` 包：WebSocket 独立标准包。
14. `websocket/servlet` 包：WebSocket 与 Servlet Upgrade 适配。
15. `tck` 包：Core Profile、注册模型、WebApp、静态资源、Session、Multipart、Async、Security、Native I/O 和 HTTP 容器兼容性测试。
16. README：说明标准定位、版本和容器兼容声明方式。

## 26. Servlet 6.1 覆盖矩阵

| Jakarta Servlet 6.1 领域 | Arkarta v0.0.1 状态 | 说明 |
| --- | --- | --- |
| Request 路径、参数、映射 | 已实现 | `RequestURI`、`QueryString`、`ContextPath`、`ServletPath`、`PathInfo`、`RequestMapping`、Parameter API |
| Response 基础与便利 API | 已实现 | Header/Status/Write/Flush/Reset、Cookie、Redirect、SendError、Content-Type、Charset、Content-Length、typed Header、Locale、Trailer |
| Servlet 路径映射 | 已实现 | exact、longest prefix、extension、default |
| RequestDispatcher | 已实现 | Forward、Include、Error 分发、完整 forward/include/error 属性族、按名称和路径获取 dispatcher |
| Servlet/Filter/Listener 注册 | 已实现 | `servlet/registration` 提供动态注册、映射冲突、初始化参数、multipart config、security constraint、冻结快照，`container` 可转换为 Deployment |
| Filter 链与 dispatcher type | 已实现 | Filter 顺序、短路、单次 `Next`、URL-pattern 约束、`ManagedFilter` 生命周期、REQUEST/FORWARD/INCLUDE/ERROR/ASYNC 位集合 |
| WebApp 生命周期和事件 | 已实现 | Context、Request、Session 生命周期监听器、属性监听器、虚拟主机名、默认字符集、MIME 映射、会话超时、资源路径、临时目录和有效规范版本 |
| Session Profile | 已实现 | Manager、MemoryManager、Accessor、Cookie 绑定、COOKIE/URL/SSL tracking policy、URL rewriting、requested ID 校验、ID 轮换、属性和绑定监听器、Store、MemoryStore、passivation/activation |
| Multipart Profile | 已实现 | 表单、Part API、请求绑定、大小限制、内存阈值、临时目录、文件名归一化、注册元数据 |
| 静态资源与 Welcome file | 已实现 | `servlet/resource` 提供 Provider、`fs.FS` 实现、default servlet、条件 GET、GET/HEAD、If-Range、弱 ETag 保护、多 Range 和 welcome file |
| Async/Stream | 已实现 | `servlet/async` 提供显式完成、Await、完成状态、dispatch 计数、超时、错误事件、ASYNC dispatch 和流式写入 |
| Upgrade/WebSocket | 已实现 | `servlet/upgrade` 提供连接交接契约和 `net/http` hijack 适配；`websocket` 提供 HTTP 握手、子协议协商、permessage-deflate 扩展、Endpoint、连接 SPI 和 TCK；`websocket/servlet` 提供 Servlet Upgrade 适配 |
| Native I/O | 已实现 | `servlet/nativeio` 提供文件区段发送契约、能力声明、发送策略、跨平台参考实现和 TCK |
| Security 声明式模型 | 已实现 | `servlet/security` 提供 Principal、Basic 认证、Realm、角色约束、方法约束、run-as、传输保障和安全 Filter；企业认证集成由 `goark-security` 承载 |
| Locale | 已实现基础 | 请求 Accept-Language 解析与响应 Content-Language 设置；i18n 资源解析由后续上层模块补充 |

## 27. 暂不解决的问题

- 是否提供 XML 描述符迁移工具。
- 与未来 `goark.dev/arkarta/web` 路由和 MVC 参数绑定的边界。
- 分布式 Session passivation/activation 的具体容器实现。
- WebSocket 具体帧读写和高性能网络容器实现。
