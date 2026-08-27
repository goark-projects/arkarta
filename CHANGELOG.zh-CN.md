# 更新日志

语言：[English](CHANGELOG.md) | 简体中文

本项目遵循 Go 模块语义化版本号。`v0.0.x` 是早期预览版本：公共 API 已按标准契约设计，但在 `v0.1.0` 前仍允许必要的源码不兼容修正。

## [0.0.1] - 2026-08-27

### 新增

- 初始化 `goark.dev/arkarta` 模块，确立 Arkarta 作为 Goark 企业级 Web 标准集合。
- 新增 Servlet Core API：`Handler`、`Servlet`、`Request`、`Response`、`Filter`、`Chain`、`WebApp`、生命周期契约、请求分发和结构化错误语义。
- 新增 Servlet 6.1 对齐能力：路径细节、参数、参数名、Cookie、映射元素、Header、Locale、Trailer、Connection 元数据、内容协商、响应 Cookie、Redirect、Error、Charset、Content-Length 和类型化 Header 辅助。
- 新增 `servlet/container` SPI，覆盖部署、应用、注册快照、生命周期、Profile 声明和容器元数据。
- 新增 `servlet/registration` 动态注册元模型，覆盖 Servlet、Filter、Listener 定义、映射、初始化参数、multipart config、security constraint、Listener 顺序和冻结快照。
- 新增基于 `net/http` 的 `servlet/nethttp` 参考适配器和最小参考容器。
- 新增 `servlet/resource` 静态资源 Provider、`fs.FS` Provider、default servlet、条件请求、弱 ETag、多 Range 响应和 welcome file 解析。
- 新增 `servlet/session` Session Profile，覆盖 Manager SPI、内存实现、Cookie/URL tracking、SSL tracking policy、requested ID 校验、ID 轮换、URL rewriting、Store SPI、passivation/activation 和监听器。
- 新增 `servlet/multipart` Multipart Profile，覆盖解析器、Part API、请求绑定、大小限制、内存阈值、临时目录、提交文件名归一化和清理语义。
- 新增 `servlet/async` Async/Stream Profile，覆盖显式完成、Await、超时、错误事件、dispatch 计数、ASYNC dispatch 和流式写入串行化。
- 新增 `servlet/upgrade` Upgrade Profile，覆盖连接交接契约和 `net/http` hijack 适配。
- 新增 `servlet/nativeio` Native I/O Profile，覆盖文件区段发送契约、能力声明、发送策略报告、跨平台 fallback 实现和 TCK。
- 新增 `servlet/security` 声明式安全 Profile，覆盖 Principal、Basic 认证、Realm、角色约束、方法约束、run-as、传输保障和安全 Filter。
- 新增 `servlet/tck` 兼容性测试，覆盖 Core HTTP、生命周期、分发、错误页、注册、WebApp 上下文、资源、Session、Multipart、Async、Security、Native I/O 和 HTTP 容器入口。
- 新增 `websocket` 标准包，覆盖 RFC 6455 握手、子协议协商、permessage-deflate 协商、端点、会话、消息、关闭状态、连接 SPI、服务循环和 JSON 文本编解码。
- 新增 `websocket/frame` RFC 6455 帧层，覆盖 Mask、扩展长度、控制帧校验、关闭帧载荷和碎片聚合。
- 新增 `websocket/servlet` 集成，覆盖 Servlet Upgrade、HTTP 101 写出、帧连接适配和 Endpoint 服务辅助。
- 新增 `websocket/tck` 兼容性测试，覆盖握手、端点生命周期、压缩和帧编解码。
- 新增 `json` 标准包，提供 `encoding/json` 默认 Codec、流式 Encoder/Decoder、输入大小限制、未知字段控制、数字精度模式和包级辅助函数。
- 新增 `json/sonic` 高性能 Codec 实现，基于 `github.com/bytedance/sonic`。
- 新增 `json/tck` Codec 兼容性测试，供标准实现和 sonic 实现共用。
- 新增 `validation` 标准包，覆盖结构体标签约束、嵌套校验、校验分组、消息解析器、对象级约束、内置约束、自定义约束和聚合错误。
- 新增 `web` MVC/REST 组合层，覆盖方法路由、路由分组、自动 HEAD/OPTIONS、路径/查询/Form/Multipart 绑定、参数转换、统一 Result、响应增强、错误映射和拦截器。
- 新增 `web/tck` 兼容性测试，覆盖 Web 路由、JSON 绑定、Validation 映射、内容协商、自动方法语义和表单绑定。
- 新增根 `security` 契约，覆盖 Principal、Authority、Authentication、Credential、AuthenticationManager、SecurityContext、Authorizer 和授权决策。
- 新增英文默认文档和简体中文镜像文档。

### 不包含

- Goark Tomcat、Goark Jetty 或其他生产级具体 Web 容器。
- 与 `jakarta.servlet` 或 `javax.servlet` 的 Java 命名空间兼容。
- JSP、JSTL、Expression Language、WAR/EAR 打包、Java 注解、运行时 classpath 扫描或 Java 式继承 API。
