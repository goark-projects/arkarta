# Arkarta Enterprise Web 1.0 标准路线

状态：Draft 1  
目标：把 Java 企业级 Web 标准体系完整复刻为 Goark 的 Go 化企业开发标准  
基线：Jakarta EE 11 Web Profile、Jakarta Servlet 6.1、Jakarta RESTful Web Services 4.0、Jakarta WebSocket 2.2、Jakarta Security 4.0、Jakarta Validation 3.1、Jakarta JSON Binding 3.0、Jakarta JSON Processing 2.1、Jakarta CDI 4.1，以及 Oracle/JCP Java EE 历史规范语义  

## 1. 总目标

Goark Enterprise Web 不是单个 HTTP 框架，而是一套企业级 Web 标准。它要复刻 Java 企业开发中已经被验证的标准边界，同时用 Go 的语言模型重建 API：

1. 底层容器契约对标 Servlet。
2. REST/MVC 对标 Jakarta RESTful Web Services 与 Spring MVC 的工程体验。
3. 安全对标 Jakarta Security 与 Spring Security 的过滤器链思想。
4. WebSocket 对标 Jakarta WebSocket。
5. 参数校验对标 Jakarta Validation。
6. JSON 编解码对标 JSON-B / JSON-P。
7. 依赖注入和生命周期对接 Goark core，而不是复制 CDI 运行时扫描。
8. 兼容性通过 Goark TCK 验证，而不是靠文档宣称。

## 2. Go 化原则

- 使用 `context.Context` 表达取消、超时、Trace 和请求作用域。
- 使用 `error` 返回表达失败，不复制 Java 异常层级。
- 使用接口组合和函数适配器，不复制 Java 类继承。
- 使用显式注册和 Goark CLI 生成注册，不复制运行时 classpath 扫描。
- 保持 `net/http` 互操作，不另造封闭生态。
- 标准层小而稳定，上层能力通过独立包和 Profile 扩展。

## 3. 分层标准

```text
Goark Enterprise Web
├── goark.dev/arkarta/servlet              底层容器契约、过滤器链、分发、会话 Profile、TCK
├── goark.dev/arkarta/web                  MVC、REST、路由、参数绑定、响应编解码
├── goark.dev/arkarta/security             认证、授权、Principal、安全过滤器
├── goark.dev/arkarta/validation           结构体验证、字段约束、错误聚合
├── goark.dev/arkarta/websocket            WebSocket 端点、消息编解码、会话管理
├── goark.dev/arkarta/json                 JSON-B / JSON-P 风格绑定与流式处理
├── goark.dev/boot                         自动装配、配置绑定、容器选择
└── 容器实现                         Goark Tomcat、Goark Jetty、net/http 容器、原生高性能容器
```

## 4. Servlet Core 负责什么

`goark.dev/arkarta/servlet` 必须先稳定以下契约：

- `Request` / `Response`
- `Handler` / `Servlet`
- `Filter` / `Chain`
- 路径映射
- `WebApp`
- 生命周期
- 错误分发
- `net/http` 适配
- TCK Core

它不直接做 Controller、参数绑定、JSON、认证授权、模板和业务路由语法。

## 5. Web/MVC 标准负责什么

未来 `goark.dev/web` 负责：

- Controller / Handler Method 模型。
- 路由变量、查询参数、Header、Cookie、Body 绑定。
- JSON、表单、Multipart、流式响应。
- 统一错误响应。
- REST 资源语义。
- 拦截器、响应建议器、异常映射器。

注册方式必须支持显式代码和 CLI 生成元数据，不依赖运行时反射扫描。

## 6. Security 标准负责什么

未来 `goark.dev/security` 负责：

- Principal 与认证结果。
- 认证过滤器链。
- 授权决策。
- Session 固定防护。
- CSRF、CORS、安全 Header。
- 方法级安全与 Web 路径安全的统一决策模型。

Servlet 层只暴露足够的请求上下文、会话 Profile 和过滤器链。

## 7. Validation 标准负责什么

未来 `goark.dev/validation` 负责：

- Struct 字段约束。
- 嵌套对象校验。
- 分组校验。
- 国际化消息。
- 统一错误聚合。

Web 层只消费校验结果并映射为 HTTP 错误。

## 8. JSON 标准负责什么

未来 `goark.dev/json` 负责：

- 结构体绑定。
- 流式 JSON 读写。
- 字段命名策略。
- 时间、枚举、数字精度策略。
- 安全解码限制。

Servlet Core 只处理请求体与响应写出，不绑定 JSON 实现。

## 9. WebSocket 标准负责什么

未来 `goark.dev/websocket` 负责：

- Endpoint 注册。
- 握手升级。
- 消息编解码。
- Ping/Pong、关闭码、超时。
- 会话属性和背压。

Servlet Upgrade Profile 只定义升级入口和连接所有权转移。

## 10. 第一阶段实现范围

当前第一阶段只实现 `goark.dev/arkarta/servlet` Core：

1. Go 模块与根包 API。
2. 路径映射与过滤器链。
3. `net/http` 适配。
4. 容器 SPI 基础类型。
5. TCK 风格测试。

上层企业标准先保留为接口路线，不在第一阶段混入实现。
