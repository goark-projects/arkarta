# Changelog

本项目遵循 Go 模块语义化版本号。`v0.0.x` 表示早期预览版本，公共 API 已按标准化方向设计，但仍允许在 `v0.1.0` 前做必要的源代码不兼容修正。

## [0.0.1] - 2026-08-26

### Added

- 初始化 `goark.dev/arkarta` 模块，确立 Arkarta 作为 Goark 企业开发标准集合。
- 新增 Servlet Core API：`Handler`、`Servlet`、`Request`、`Response`、`Filter`、`Chain`、`WebApp`、生命周期和错误模型。
- 新增 Servlet 6 对齐能力：请求路径/参数/映射元素、响应 Cookie/Redirect/Error/Charset 便利 API、Filter 生命周期和 dispatcher type 过滤。
- 新增 `servlet/container` 容器 SPI：`Container`、`Application`、`Deployment`、路径映射和 Profile 元数据。
- 新增 `servlet/registration` 动态注册元模型，覆盖 Servlet、Filter、Listener 注册、初始化参数、URL 映射和冻结快照。
- 新增 `servlet/nethttp` 参考容器和 `net/http` 适配器。
- 新增 `servlet/session` Session Profile 接口、请求/响应 Cookie 绑定和内存会话管理器。
- 新增 `servlet/multipart` Multipart Profile 解析器。
- 新增 `servlet/tck` 兼容性测试工具，覆盖 Core HTTP、生命周期、分发、错误页、注册元模型、Session、Multipart 和 HTTP 容器入口。
- 新增 Arkarta Servlet 与 Enterprise Web 标准草案文档。

### Not Included

- 尚未发布 Goark Tomcat、Goark Jetty 等独立具体容器。
- 尚未实现 Async/Stream Profile、Upgrade Profile 和 Native I/O Profile。
- 尚未实现上层 MVC、REST、Security、Validation、JSON 和 WebSocket 标准包。
