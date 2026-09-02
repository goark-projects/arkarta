# ADR-0001：Servlet 标准契约保持传输层中立

## 状态

已接受

## 日期

2026-09-02

## 背景

Arkarta 的定位与 Jakarta EE Servlet API 相同：定义应用、容器和扩展之间的公共标准，并通过 TCK 证明实现兼容性。Arkhos 以及其他第三方容器的定位与 Tomcat、Jetty 相同：实现 Arkarta 标准，并可由应用在不修改业务代码的前提下替换。

现有 Servlet 请求、响应、Multipart、Basic 认证和 Upgrade 契约直接暴露了 `net/http` 类型。这会产生以下问题：

- 应用代码能够依赖某个容器的传输实现，破坏容器可替换性。
- Hertz 等实现必须在热路径构造 `http.Request` 和 `http.ResponseWriter` 兼容对象。
- Async 请求越过底层对象池生命周期时，所有权和复制边界不明确。
- TCK 只能验证 `net/http` 适配行为，不能独立验证标准语义。

## 决策

Arkarta Servlet 公共契约必须保持传输层中立：

1. `servlet` 及其标准 Profile 不得在公共 API 中暴露 `*http.Request`、`http.ResponseWriter`、`http.Header`、`http.Cookie` 或底层框架上下文。
2. Arkarta 定义自己的 Request、Response、Header、Cookie、Body、Trailer、Multipart、Async 和 Upgrade 契约。
3. 请求采用只读元数据端口与标准管理状态分离的设计；属性、分发、映射和异步状态由标准语义管理。
4. 响应采用小接口组合，Header、状态、正文、Flush、Reset、Trailer 和 Upgrade 能力分别由消费方按需检测。
5. `servlet/nethttp` 只负责 `net/http` 与 Arkarta 契约之间的适配，不得成为标准核心的反向依赖。
6. TCK 只依赖 Arkarta 契约和容器测试驱动，不依赖某个 HTTP 框架；传输实现可增加自己的互操作测试。
7. 容器特有扩展只能位于实现仓库或明确的扩展包，不能进入标准核心接口。

## Java EE 语义映射

| Jakarta Servlet 概念 | Arkarta Go 契约 |
|---|---|
| `ServletRequest` / `HttpServletRequest` | `servlet.Request` |
| `ServletResponse` / `HttpServletResponse` | `servlet.Response` |
| `Cookie` | `servlet.Cookie` |
| `AsyncContext` | `servlet/async.Context` |
| `Part` | `servlet/multipart.Part` |
| `HttpUpgradeHandler` | `servlet/upgrade.Handler` |
| Servlet Container SPI | `servlet/container.Container` |
| Servlet TCK | `servlet/tck` |

映射只对齐职责、生命周期和可替换性，不复制 Java 的继承体系、异常模型或运行时扫描机制。

## 备选方案

### 保留 `net/http` 作为标准数据模型

实现简单，但会永久绑定 Go 标准库对象模型，并让直接 Hertz 实现失去意义，因此拒绝。

### 在 Hertz 前增加 `net/http` 转换层

适合作为短期原型，但会增加对象构造、Header 转换和 Body 包装成本，也无法自然表达池化对象所有权，因此拒绝作为生产路径。

### 为每种传输复制一套 Servlet API

会导致应用 API 分裂和 TCK 重复，违反单一标准原则，因此拒绝。

## 后果

- 现有 `v0.x` 公共 API 会发生一次有意的破坏性收敛。
- `servlet/nethttp`、Arkhos Hertz、Goark Web 和相关 starter 必须迁移到同一标准契约。
- 容器实现可以针对同步热路径进行零拷贝或延迟复制，并在 Async 边界显式快照。
- 任何新的容器实现只需实现 Arkarta 契约并通过同一套 TCK。
