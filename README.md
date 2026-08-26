# Arkarta

Arkarta 是 Goark 的企业开发标准模块。它参考 Jakarta EE / Java EE 的成熟企业标准体系，但 API 和运行时边界必须 Go 化：显式注册、接口组合、`context.Context`、`net/http` 互操作、错误返回和可验证的 TCK。

当前阶段先完成 Servlet 部分。Servlet 相关代码全部位于 `servlet` 包及其子包下，后续 Web、Security、Validation、JSON、WebSocket 等标准也会各自拆包，保持职责边界清晰。

## 当前版本

当前首个预览版本为 `v0.0.1`。该版本用于固定 Arkarta Servlet Core Profile 的第一批公共契约，并给后续 Goark Tomcat、Goark Jetty 等具体 Web 容器提供可执行的 TCK 基线。

`v0.0.1` 包含：

- Servlet Core API：请求、响应、处理器、Servlet、过滤器链、WebApp、生命周期、错误模型和请求分发。
- Servlet 6 对齐能力：请求路径/参数/映射元素、响应 Cookie/Redirect/Error/Charset 便利 API、Filter 生命周期与 dispatcher type 过滤、静态资源、welcome file 和 Session URL rewriting。
- Servlet Container SPI：部署描述、注册快照转换、应用生命周期、Profile 声明和容器元数据。
- `servlet/registration`：Servlet、Filter、Listener 动态注册元模型。
- `servlet/nethttp`：基于标准库 `net/http` 的参考适配和最小参考容器。
- `servlet/resource`：静态资源 Provider、`fs.FS` 实现、default servlet 和 welcome file 解析。
- `servlet/session`：Session Profile 接口、请求/响应 Cookie 绑定、URL rewriting 和内存会话管理器。
- `servlet/multipart`：Multipart Profile 解析器。
- `servlet/tck`：Core HTTP、生命周期、分发、错误页、注册元模型、Session、Multipart 和 HTTP 容器入口的兼容性测试。

`v0.0.1` 不包含：

- Goark Tomcat、Goark Jetty 等独立具体容器。
- Async/Stream Profile、Upgrade Profile、Native I/O Profile。
- 上层 MVC、REST、Security、Validation、JSON、WebSocket 标准包。

## 安装

```shell
go get goark.dev/arkarta@v0.0.1
```

当前包结构：

```text
servlet/            Servlet Core API
servlet/container  Servlet 容器 SPI
servlet/multipart  Multipart Profile
servlet/nethttp    net/http 适配
servlet/registration 动态注册元模型
servlet/resource   静态资源与 default servlet
servlet/session    Session Profile
servlet/tck        Servlet 兼容性测试工具
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

## 许可证

Arkarta 使用 [Apache License 2.0](LICENSE) 发布。
