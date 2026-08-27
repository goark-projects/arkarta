# Arkarta Enterprise Web 1.0

语言：[English](arkarta-enterprise-web-1.0.md) | 简体中文

状态：Release Candidate 1

模块：`goark.dev/arkarta`

目标版本：`v0.0.1`

日期：2026-08-27

## 1. 目标

Arkarta Enterprise Web 是 Goark 的企业级 Web 标准。它不是单个 HTTP 框架，而是定义应用、Web 容器、路由层、校验层、JSON 编解码、WebSocket 端点和安全集成之间的互操作契约。

参考模型来自 Java 企业标准体系：Jakarta Servlet、Jakarta RESTful Web Services、Jakarta WebSocket、Jakarta Security、Jakarta Validation、JSON-B、JSON-P，以及 Java EE / Oracle 历史规范。Arkarta 不复制 Java API，而是保留成熟企业边界，并用 Go 化方式重新表达。

## 2. Go 化规则

- 使用 `context.Context` 表达取消、截止时间、Trace 传播和请求作用域。
- 使用显式 `error` 返回，不复制 Java 异常层级。
- 使用小接口和函数适配器，不使用继承模型。
- 使用显式注册或生成注册，不使用运行时 classpath 扫描。
- 保持 `net/http` 互操作。
- 标准按包拆分，并能独立测试。
- 兼容性声明必须有 TCK 支撑。

## 3. 标准分层

```text
goark.dev/arkarta/servlet              Servlet Core 与容器契约
goark.dev/arkarta/servlet/*            Servlet Profile 与适配
goark.dev/arkarta/web                  MVC/REST 组合层
goark.dev/arkarta/json                 JSON Codec 标准与 encoding/json 实现
goark.dev/arkarta/json/sonic           Sonic JSON 实现
goark.dev/arkarta/validation           Validation 标准
goark.dev/arkarta/websocket            WebSocket 标准
goark.dev/arkarta/websocket/frame      RFC 6455 帧层
goark.dev/arkarta/websocket/servlet    WebSocket 的 Servlet Upgrade 适配
goark.dev/arkarta/security             根企业安全契约
goark.dev/arkarta/*/tck                可执行兼容性测试
```

Goark Tomcat、Goark Jetty 等具体容器实现这些契约，但不属于 Arkarta 标准模块本身。

## 4. Servlet 职责

`goark.dev/arkarta/servlet` 定义底层应用/容器边界：

- 请求和响应模型。
- Handler 与 Servlet 契约。
- Filter 链与 dispatcher type。
- WebApp 上下文和生命周期。
- 路径映射与分发。
- 错误模型和错误页。
- `net/http` 互操作。
- 可选 Profile：Session、Multipart、Async/Stream、Upgrade、Native I/O 和 Servlet 声明式安全。
- Servlet TCK 入口。

Servlet 不负责 Controller 方法绑定、JSON 编码、企业认证提供者、模板或业务路由语法。

## 5. Web 职责

`goark.dev/arkarta/web` 定义 Goark MVC/REST 组合层：

- 带路径变量的方法路由。
- 路由分组和分组级拦截器。
- `GET` 自动支持 `HEAD`，并自动提供带稳定 `Allow` 的 `OPTIONS`。
- Query、Path、Header、Cookie、Form、Multipart 绑定辅助。
- 参数转换辅助。
- JSON 绑定和 Validation 集成。
- JSON、Text、NoContent 统一 `Result` 模型。
- Response Advice。
- 错误映射。
- 容器集成 Web TCK。

Controller 方法发现和生成注册属于 Goark 工具，不属于标准运行时反射。

## 6. JSON 职责

`goark.dev/arkarta/json` 定义参考 JSON-B / JSON-P 的 Codec 契约：

- `Codec`、`Encoder`、`Decoder` 接口。
- `encoding/json` 默认实现。
- 流式编码/解码。
- 输入大小限制。
- 未知字段拒绝。
- 数字精度模式。
- HTML 转义和缩进控制。

`goark.dev/arkarta/json/sonic` 是标准高性能实现。`goark.dev/arkarta/json/tck` 用于验证 Codec 兼容性。

## 7. Validation 职责

`goark.dev/arkarta/validation` 定义 Go 化校验：

- 结构体标签字段约束。
- 嵌套结构体和切片校验。
- 通过 `arkarta-groups` 支持校验分组。
- 消息解析器扩展点。
- 对象级约束。
- 内置约束：`required`、`notblank`、`min`、`max`、`len`、`email`、`oneof`、`regexp`、`url`、`uuid`、`gt`、`gte`、`lt`、`lte`、`contains`、`startswith`、`endswith`。
- 自定义约束注册。
- 聚合校验结果和稳定校验错误契约。

标准刻意避免 Java 注解扫描。约束通过显式 Go 值或结构体标签表达。

## 8. WebSocket 职责

`goark.dev/arkarta/websocket` 定义 WebSocket 标准：

- HTTP Upgrade 握手。
- 子协议协商。
- permessage-deflate 扩展协商。
- Endpoint 和 Session 契约。
- Message 和 Close Status 模型。
- Connection SPI 和服务循环。
- JSON 文本 Codec。
- WebSocket TCK。

`websocket/frame` 负责 RFC 6455 帧编解码。`websocket/servlet` 负责把 Servlet Upgrade 连接适配为 WebSocket Session。

## 9. Security 职责

Security 刻意拆分：

- `servlet/security` 定义 Servlet 声明式安全：请求绑定 Principal、Basic 认证、Realm、角色约束、方法约束、run-as、传输保障和安全 Filter。
- 根 `security` 定义与 Servlet 无关的企业安全契约：Principal、Authority、Authentication、Credential、AuthenticationManager、SecurityContext、Authorizer 和授权决策。

具体身份提供者、OAuth2/OIDC/JWT、CSRF、CORS、密码存储和策略引擎由标准之上的安全实现承载。

## 10. TCK 策略

模块或容器只能声明自己已经通过的 Arkarta Profile。相关 TCK 包括：

- `servlet/tck`：Servlet Core 和 Profile。
- `web/tck`：基于容器的 Web、JSON、Validation 集成。
- `websocket/tck`：WebSocket 握手、端点、压缩和帧行为。
- `json/tck`：JSON Codec 实现。

本仓库发布门禁：

```shell
go test ./...
go test -race ./...
go vet ./...
gofmt -l .
```

`gofmt -l .` 必须无输出。

## 11. 版本策略

`v0.0.1` 固定第一版公开预览标准。在 `v0.1.0` 前，如果标准质量需要，仍允许源码不兼容修正。进入 `v0.1.0` 后，公共 API 应尽量采用加法演进。
