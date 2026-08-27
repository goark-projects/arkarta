# Arkarta Servlet 1.0

语言：[English](arkarta-servlet-1.0.md) | 简体中文

状态：Release Candidate 1

模块：`goark.dev/arkarta/servlet`

目标版本：`v0.0.1`

日期：2026-08-27

## 1. 目的

Arkarta Servlet 定义 Goark Web 应用与 Web 容器之间的稳定边界。它以 Jakarta Servlet 6.1 作为主要参考标准，并参考 Java Servlet / Oracle Java EE 历史语义，但不声明 Jakarta 兼容，也不复用 `jakarta.servlet` 或 `javax.servlet` 命名空间。

Arkarta 容器只有实现对应契约并通过对应 Arkarta TCK，才能声明兼容。

## 2. 规范用语

- 必须：兼容性强制要求。
- 应当：默认期望，除非存在明确工程理由。
- 可以：可选行为。
- 容器：负责部署应用、管理生命周期、分发请求并暴露支持 Profile 的运行时。
- 应用：注册到容器中的用户代码。

## 3. 设计原则

- Go 优先：使用 `context.Context`、`error`、接口、函数和 `net/http`。
- 容器中立：不假设 Tomcat、Jetty、`net/http`、epoll、kqueue、io_uring 或任何事件循环。
- 小核心、显式 Profile：Session、Multipart、Async、Upgrade、Native I/O、Security 和 WebSocket 集成都拆到独立包。
- 不使用 Java 继承模型：不提供 `GenericServlet`、`HttpServlet`、注解扫描、WAR/EAR 模型或 classloader 语义。
- 兼容性可执行：必须行为需要由 TCK 覆盖。

## 4. Core Profile

每个 Arkarta Servlet 容器必须支持：

- `Handler`、`HandlerFunc` 和带生命周期的 `Servlet`。
- `Request` 和 `Response`。
- `Filter`、`ManagedFilter` 和 `Chain`。
- `WebApp` 上下文、初始化参数、属性、资源、监听器、MIME 映射、临时目录、会话超时和版本元数据。
- Servlet 路径映射：精确、最长前缀、扩展名和默认映射。
- Dispatch type：request、forward、include、error、async。
- RequestDispatcher forward/include/error 行为。
- Status error、panic 恢复、错误页和已提交响应安全性。
- `net/http` 互操作。

## 5. Request 契约

`Request` 必须暴露：

- Method、protocol、scheme、host、remote address、secure 标记和 connection ID。
- `RequestURI`、`RequestURL`、`QueryString`、`ContextPath`、`Path`、`ServletPath`、`PathInfo`、`RequestMapping`。
- Query 值，以及 query 加 `application/x-www-form-urlencoded` body 的合并 Servlet 参数视图。
- 稳定排序的参数名。
- Header、类型化 Header 辅助、Cookie 和 Cookie 快照。
- 从 `Accept-Language` 解析 Locale。
- Accept 媒体类型解析和内容协商辅助。
- Trailer 字段和原始 `*http.Request` 互操作。
- Body reader 和请求属性。

Query 值必须只表示查询串。Servlet parameters 才是合并参数视图。

## 6. Response 契约

`Response` 必须暴露：

- Header、status、body write、string write、flush、committed flag、reset 和 body writer。
- Cookie 辅助。
- Redirect 和 error 辅助。
- Content-Type、charset、content length、typed header、locale 和 trailer 辅助。

第一次写 body 必须提交响应。提交后 `Reset` 必须失败。默认错误处理不得泄露堆栈或本地路径。

## 7. Filter 与 Dispatch 契约

容器必须提供确定的 Filter 顺序。Filter 可以通过不调用 `chain.Next` 短路，但不得多次调用 `Next`。Filter binding 必须遵守 dispatcher type 和 URL-pattern 匹配。

Forward 必须发生在响应提交前。Include 不得覆盖主响应状态。Error dispatch 必须暴露稳定请求属性，包括 status、exception、request URI、servlet name 和 message。

## 8. v0.0.1 已实现 Optional Profile

### Session

`servlet/session` 定义 Session manager 和 accessor 契约、Cookie/URL tracking、SSL tracking policy、ID 轮换、requested ID 校验、URL rewriting、Store SPI、MemoryStore、passivation/activation 和监听器。

### Multipart

`servlet/multipart` 定义解析器选项、最大请求体、最大文件、内存阈值、临时目录、提交文件名归一化、`Part`、请求绑定和清理。

### Async 与 Stream

`servlet/async` 定义显式异步生命周期、完成、超时、错误、dispatch 计数、await 和流式写入。Go goroutine 本身不等于 Servlet async 生命周期，应用必须显式进入该 Profile。

### Upgrade

`servlet/upgrade` 定义连接所有权转移和 `net/http` hijack 适配。WebSocket 协议行为由 `goark.dev/arkarta/websocket` 承载。

### Native I/O

`servlet/nativeio` 定义文件区段发送、能力声明、发送策略报告、跨平台 fallback sender、非法区段处理和取消行为。

### 声明式安全

`servlet/security` 定义请求绑定 Principal、Basic authenticator、Realm、角色约束、方法约束、角色映射、run-as、传输保障和安全 Filter。

## 9. Container SPI

`servlet/container` 定义：

- `Container`
- `Application`
- `Deployment`
- 注册快照转换
- 应用生命周期
- Profile 声明
- 容器名称、版本、支持 Profile 和限制参数元数据

容器必须在部署阶段校验映射冲突、注册冻结规则、生命周期顺序和 Profile 依赖。

## 10. Servlet 6.1 覆盖矩阵

| Servlet 领域 | Arkarta v0.0.1 状态 |
| --- | --- |
| Request 路径与映射细节 | 已实现 |
| Query/form 参数与参数名 | 已实现 |
| Header、Locale、Cookie、Trailer、Connection 元数据 | 已实现 |
| Accept 协商与响应 Content-Type 辅助 | 已实现 |
| Response 状态、Header、Body、Cookie、Redirect、Error、Charset、Trailer | 已实现 |
| Servlet 路径映射 | 已实现 |
| RequestDispatcher forward/include/error | 已实现 |
| Servlet、Filter、Listener 动态注册 | 已实现 |
| Filter 生命周期与 dispatcher type 过滤 | 已实现 |
| WebApp 上下文与监听器 | 已实现 |
| Session Profile | 已实现 |
| Multipart Profile | 已实现 |
| 静态资源与 Welcome file | 已实现 |
| Async 与 Stream Profile | 已实现 |
| Upgrade Profile | 已实现 |
| WebSocket Adapter | 通过 `websocket/servlet` 实现 |
| Native I/O Profile | 已实现 |
| 声明式安全 Profile | 已实现 |
| Java 注解扫描与 WAR/EAR 打包 | 不包含 |
| JSP、JSTL、EL、Java classloader 语义 | 不包含 |
| HTTP/2 server push 强制要求 | 不包含 |

## 11. 兼容性

容器必须运行与声明 Profile 对应的 TCK 入口：

- `servlet/tck.RunCoreHTTP`
- `servlet/tck.RunLifecycle`
- `servlet/tck.RunErrorPages`
- `servlet/tck.RunSessionManager`
- `servlet/tck.RunSessionRequestBinding`
- `servlet/tck.RunMultipartParser`
- `servlet/tck.RunAsyncLifecycle`
- `servlet/tck.RunSecurity`
- `servlet/tck.RunNativeIO`
- `servlet/tck.RunStaticResources`
- `servlet/tck.RunHTTPContainer`

通过 Core 不代表通过 Optional Profile。

## 12. 版本策略

预览期后 Arkarta Servlet 1.x 应尽量采用加法演进。可选能力应通过新包、小接口或 option 扩展，而不是修改已发布方法签名。
