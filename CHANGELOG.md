# Changelog

本项目遵循 Go 模块语义化版本号。`v0.0.x` 表示早期预览版本，公共 API 已按标准化方向设计，但仍允许在 `v0.1.0` 前做必要的源代码不兼容修正。

## [Unreleased]

### Added

- Servlet Core 新增 Accept 媒体类型解析、质量因子排序、通配符匹配和响应 Content-Type 协商辅助。

## [0.0.1] - 2026-08-27

### Added

- 初始化 `goark.dev/arkarta` 模块，确立 Arkarta 作为 Goark 企业开发标准集合。
- 新增 Servlet Core API：`Handler`、`Servlet`、`Request`、`Response`、`Filter`、`Chain`、`WebApp`、生命周期和错误模型。
- 新增 Servlet 6.1 对齐能力：请求路径/参数/映射元素、请求 Header/Locale/Trailer/Connection 元数据、响应 Cookie/Redirect/Error/Charset/Header/Trailer 便利 API、Filter 生命周期、dispatcher type 过滤、静态资源 Range、welcome file 和 Session tracking。
- 新增 `servlet/container` 容器 SPI：`Container`、`Application`、`Deployment`、注册快照转换、路径映射和 Profile 元数据。
- 新增 `servlet/registration` 动态注册元模型，覆盖 Servlet、Filter、Listener 注册、初始化参数、URL 映射、multipart config、security constraint 和冻结快照。
- 新增 `servlet/nethttp` 参考容器和 `net/http` 适配器。
- 新增 `servlet/resource` 静态资源 Provider、`fs.FS` 实现、default servlet 和 welcome file 解析。
- 新增 `servlet/session` Session Profile 接口、请求/响应 Cookie 绑定、COOKIE/URL/SSL tracking policy、URL rewriting、属性监听和内存会话管理器。
- 新增 `servlet/multipart` Multipart Profile 解析器、Part API 和请求绑定。
- 新增 `servlet/async` Async/Stream Profile，覆盖显式完成、超时、错误事件、ASYNC dispatch 和流式写入。
- 新增 `servlet/upgrade` Upgrade Profile，覆盖 HTTP 升级连接交接契约和 `net/http` hijack 适配。
- 新增 `servlet/nativeio` Native I/O Profile，覆盖文件区段发送、能力声明、发送策略和跨平台参考实现。
- 新增 `servlet/security` 声明式安全 Profile，覆盖 Principal、Basic 认证、Realm、角色约束、方法约束、run-as、传输保障和安全 Filter。
- 新增 `servlet/tck` 兼容性测试工具，覆盖 Core HTTP、生命周期、分发、错误页、注册元模型、WebApp 上下文能力、静态资源、Session、Multipart、Async、Security、Native I/O 和 HTTP 容器入口。
- 新增 `websocket` 独立标准包，覆盖 HTTP 握手、子协议协商、permessage-deflate 扩展、端点、会话、消息、关闭原因、连接 SPI、服务循环和 JSON 文本编解码。
- 新增 `websocket/servlet` 适配包，覆盖 WebSocket 握手、Servlet Upgrade 连接移交和 HTTP 101 写出。
- 新增 `websocket/tck` 兼容性测试工具，覆盖 HTTP 握手、子协议顺序、扩展协商、permessage-deflate 往返与限制、HTTP 101 写出和 Endpoint 生命周期。
- 新增 Arkarta Servlet 与 Enterprise Web 标准草案文档。

### Hardened

- 注册模型补齐冻结快照、注册关闭、load-on-startup 顺序和同名 Servlet 单次初始化语义。
- Async/Stream Profile 补齐 Await、完成状态、dispatch 计数、超时事件顺序和流式写入串行化语义。
- Session Profile 补齐 WebApp 级 CookieConfig、SSL tracking、Store SPI、MemoryStore、passivation/activation 和属性名快照。
- Multipart Profile 补齐临时目录落盘、提交文件名归一化、Part 删除和表单清理语义。
- 静态资源与错误页补齐 If-Range、弱 ETag、多 range、默认错误页、错误页循环保护和错误类型后注册优先语义。

### Not Included

- 尚未发布 Goark Tomcat、Goark Jetty 等独立具体容器。
- 尚未实现上层 MVC、REST、Validation、JSON 和企业级安全集成标准包。
