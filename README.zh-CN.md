# Arkarta

语言：[English](README.md) | 简体中文

Arkarta 是 Goark 的企业级 Web 应用与 Web 容器标准。它参考 Jakarta EE / Java EE 已经被长期验证的标准边界，但 API 采用 Go 化表达：显式注册、小接口、`context.Context`、`net/http` 互操作、错误返回和可执行 TCK。

第一版候选版本聚焦 Servlet、WebSocket、JSON、Validation、Web 路由/绑定和安全标准契约。Goark Tomcat、Goark Jetty 等具体容器不在本仓库实现，它们后续作为独立容器实现 Arkarta 标准。

## 状态

当前目标：`v0.0.1` release candidate。

旧 tag 已删除，下面命令只应在重新创建并发布 `v0.0.1` tag 后使用：

```shell
go get goark.dev/arkarta@v0.0.1
```

本地开发验证：

```shell
go test ./...
go test -race ./...
go vet ./...
```

## 包结构

| 包 | 职责 |
| --- | --- |
| `servlet` | 请求、响应、处理器、Servlet、过滤器、分发、WebApp、生命周期、错误和 `net/http` 互操作核心契约。 |
| `servlet/container` | 容器 SPI、部署元数据、应用生命周期、注册快照、Profile 声明和容器元数据。 |
| `servlet/registration` | Servlet、Filter、Listener 动态注册模型，覆盖初始化参数、映射、multipart config、security constraint 和冻结快照。 |
| `servlet/nethttp` | 标准库 `net/http` 适配和最小参考容器。 |
| `servlet/resource` | 静态资源 Provider、`fs.FS` Provider、default servlet、条件请求、Range 和 welcome file。 |
| `servlet/session` | Session Profile、Cookie/URL tracking、ID 轮换、请求绑定、Store SPI、内存实现和监听器。 |
| `servlet/multipart` | Multipart Profile、解析器、Part API、上传限制、临时文件和请求绑定。 |
| `servlet/async` | Go 化 Async/Stream Profile，覆盖完成、超时、错误、dispatch 和流生命周期。 |
| `servlet/upgrade` | 协议升级 Profile 和 `net/http` hijack 适配。 |
| `servlet/nativeio` | Native I/O Profile，覆盖文件区段、发送策略、能力声明和跨平台 fallback sender。 |
| `servlet/security` | Servlet 声明式安全 Profile，覆盖 Principal、Basic 认证、Realm、角色约束、方法约束、run-as 和安全 Filter。 |
| `servlet/tck` | Servlet 容器 Core 与各 Profile 兼容性测试。 |
| `websocket` | WebSocket 标准：握手、子协议协商、permessage-deflate、端点、会话、消息、关闭状态、连接 SPI、服务循环和 JSON 文本编解码。 |
| `websocket/frame` | RFC 6455 帧读写、Mask、扩展长度、控制帧校验、关闭帧载荷和碎片聚合。 |
| `websocket/servlet` | Servlet Upgrade 适配、HTTP 101 写出、帧连接适配和 Endpoint 服务辅助。 |
| `websocket/tck` | WebSocket 握手、端点生命周期、压缩和帧层兼容性测试。 |
| `json` | JSON Codec 标准，提供 `encoding/json` 默认实现、流式 API、大小限制、未知字段控制和数字精度模式。 |
| `json/sonic` | 基于 sonic 的高性能 JSON 实现，并遵守 Arkarta JSON 契约。 |
| `json/tck` | JSON Codec 兼容性测试。 |
| `validation` | Go 化 Validation 标准，覆盖结构体标签、分组、消息解析、对象级约束、嵌套校验、内置约束和自定义约束。 |
| `web` | MVC/REST 组合层，覆盖方法路由、路由分组、自动 HEAD/OPTIONS、路径/查询/Form/Multipart 绑定、响应结果、响应增强、错误映射和拦截器。 |
| `web/tck` | Web、JSON、Validation 组合兼容性测试。 |
| `security` | 根企业安全契约：Principal、Authority、Authentication、Credential、AuthenticationManager、SecurityContext、Authorizer 和授权决策。 |

## 设计规则

- Arkarta 不是 Java API 逐字翻译，而是在保留企业语义的前提下采用 Go 语言习惯。
- 标准按包边界拆分。Servlet 相关 API 位于 `servlet` 及其子包；根 `security`、`json`、`validation`、`web`、`websocket` 作为独立标准存在。
- 避免运行时扫描、Java 式代理和继承式模型。注册必须显式完成，或由 Goark 工具生成。
- 兼容性声明必须有对应 TCK 结果支撑，不能只靠文档描述。
- Optional Profile 独立声明、独立测试；容器只能声明自己已经通过的 Profile。

## 文档

- [Arkarta Enterprise Web 1.0](docs/spec/arkarta-enterprise-web-1.0.zh-CN.md)
- [Arkarta Servlet 1.0](docs/spec/arkarta-servlet-1.0.zh-CN.md)
- [Servlet 容器 TCK 接入指南](docs/tck/servlet-container.zh-CN.md)
- [更新日志](CHANGELOG.zh-CN.md)

英文默认文档：

- [Arkarta Enterprise Web 1.0](docs/spec/arkarta-enterprise-web-1.0.md)
- [Arkarta Servlet 1.0](docs/spec/arkarta-servlet-1.0.md)
- [Servlet Container TCK Guide](docs/tck/servlet-container.md)
- [Changelog](CHANGELOG.md)

## 许可证

Arkarta 使用 [Apache License 2.0](LICENSE) 发布。
