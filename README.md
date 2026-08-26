# Arkarta

Arkarta 是 Goark 的企业开发标准模块。它参考 Jakarta EE / Java EE 的成熟企业标准体系，但 API 和运行时边界必须 Go 化：显式注册、接口组合、`context.Context`、`net/http` 互操作、错误返回和可验证的 TCK。

当前阶段先完成 Servlet 部分。Servlet 相关代码全部位于 `servlet` 包及其子包下，后续 Web、Security、Validation、JSON、WebSocket 等标准也会各自拆包，保持职责边界清晰。

当前包结构：

```text
servlet/            Servlet Core API
servlet/container  Servlet 容器 SPI
servlet/multipart  Multipart Profile
servlet/nethttp    net/http 适配
servlet/session    Session Profile
servlet/tck        Servlet 兼容性测试工具
```

标准文档：

- [Arkarta Enterprise Web 1.0 标准路线](docs/spec/arkarta-enterprise-web-1.0.md)
- [Arkarta Servlet 1.0 规范草案](docs/spec/arkarta-servlet-1.0.md)
