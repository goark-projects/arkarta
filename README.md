# Arkarta

Arkarta 是 Goark 的企业开发标准模块。它参考 Jakarta EE / Java EE 的成熟企业标准体系，但 API 和运行时边界必须 Go 化：显式注册、接口组合、`context.Context`、`net/http` 互操作、错误返回和可验证的 TCK。

当前阶段先完成 Servlet 部分，并提供 WebSocket、JSON、Validation、Web 第一版独立标准包。Servlet 相关代码全部位于 `servlet` 包及其子包下，WebSocket 位于根级 `websocket` 包；后续上层安全集成等标准也会各自拆包，保持职责边界清晰。

## 当前版本

当前首个预览版本为 `v0.0.1`。该版本用于固定 Arkarta Servlet Core Profile 的第一批公共契约，并给后续 Goark Tomcat、Goark Jetty 等具体 Web 容器提供可执行的 TCK 基线。

`v0.0.1` 包含：

- Servlet Core API：请求、响应、处理器、Servlet、过滤器链、WebApp、生命周期、错误模型和请求分发。
- Servlet 6.1 对齐能力：请求路径/参数/映射元素、Header/Locale/Trailer/Connection 元数据、响应 Cookie/Redirect/Error/Charset/Trailer 便利 API、Filter 生命周期、dispatcher type 过滤、静态资源、Range、welcome file 和 Session tracking。
- Servlet Container SPI：部署描述、注册快照转换、应用生命周期、Profile 声明和容器元数据。
- `servlet/registration`：Servlet、Filter、Listener 动态注册元模型、multipart config 和 security constraint 元数据。
- `servlet/nethttp`：基于标准库 `net/http` 的参考适配和最小参考容器。
- `servlet/resource`：静态资源 Provider、`fs.FS` 实现、default servlet 和 welcome file 解析。
- `servlet/session`：Session Profile 接口、请求/响应 Cookie 绑定、COOKIE/URL/SSL tracking policy、URL rewriting、属性监听和内存会话管理器。
- `servlet/multipart`：Multipart Profile 解析器、Part API 和请求绑定。
- `servlet/async`：Go 化 Async/Stream Profile，提供显式完成、超时、错误事件和流式写入。
- `servlet/upgrade`：协议升级 Profile，定义连接交接契约和 `net/http` hijack 适配。
- `servlet/nativeio`：Native I/O Profile，定义文件区段发送、能力声明、发送策略和跨平台参考实现。
- `servlet/security`：声明式安全 Profile，提供 Principal、Basic 认证、Realm、角色约束、方法约束、run-as 和安全 Filter。
- `servlet/tck`：Core HTTP、生命周期、分发、错误页、注册元模型、WebApp 上下文能力、静态资源、Session、Multipart、Async、Security、Native I/O 和 HTTP 容器入口的兼容性测试。
- `websocket`：独立 WebSocket 标准包，覆盖 HTTP 握手、Servlet Upgrade 适配、子协议协商、permessage-deflate 扩展、端点、会话、消息、关闭原因、连接 SPI、服务循环、JSON 文本编解码和 TCK。
- `json`：JSON-B / JSON-P 风格标准入口，提供 `encoding/json` 默认实现和 `json/sonic` 高性能实现。
- `validation`：Jakarta Validation 风格的 Go 化校验入口，提供结构体标签约束、嵌套对象校验、错误聚合和自定义约束扩展点。
- `web`：MVC/REST 组合层第一版，提供方法路由、路径变量、请求绑定、统一 Result、错误响应映射和拦截器链。

`v0.0.1` 不包含：

- Goark Tomcat、Goark Jetty 等独立具体容器。
- 企业级安全集成标准包。

## 安装

```shell
go get goark.dev/arkarta@v0.0.1
```

当前包结构：

```text
servlet/            Servlet Core API
servlet/container  Servlet 容器 SPI
servlet/async      Async/Stream Profile
servlet/multipart  Multipart Profile
servlet/nativeio   Native I/O Profile
servlet/nethttp    net/http 适配
servlet/registration 动态注册元模型
servlet/resource   静态资源与 default servlet
servlet/security   声明式安全 Profile
servlet/session    Session Profile
servlet/tck        Servlet 兼容性测试工具
servlet/upgrade    协议升级 Profile
websocket          WebSocket 标准包
websocket/servlet  Servlet Upgrade 适配
websocket/tck      WebSocket 兼容性测试工具
json               JSON 标准接口与 encoding/json 实现
json/sonic         sonic 高性能 JSON 实现
validation         结构体参数校验标准
web                MVC/REST 组合层标准
```

## 验证

```shell
go test ./...
go test -race ./...
go vet ./...
```

标准文档：

- [Arkarta Enterprise Web 1.0 标准路线](docs/spec/arkarta-enterprise-web-1.0.md)
- [Arkarta Servlet 1.0 规范草案](docs/spec/arkarta-servlet-1.0.md)
- [Arkarta Servlet TCK 接入指南](docs/tck/servlet-container.md)

## 许可证

Arkarta 使用 [Apache License 2.0](LICENSE) 发布。
