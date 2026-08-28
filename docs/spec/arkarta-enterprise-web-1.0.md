# Arkarta Enterprise Web 1.0

Language: English | [简体中文](arkarta-enterprise-web-1.0.zh-CN.md)

Status: Published

Module: `goark.dev/arkarta`

Version: `v0.0.1`

Date: 2026-08-27

## 1. Goal

Arkarta Enterprise Web is the Goark enterprise Web standard. It is not a single HTTP framework. It defines the contracts that applications, Web containers, routing layers, validation layers, JSON codecs, WebSocket endpoints, and security integrations use to interoperate.

The reference model is the Java enterprise standard family: Jakarta Servlet, Jakarta RESTful Web Services, Jakarta WebSocket, Jakarta Security, Jakarta Validation, JSON-B, JSON-P, and the historical Java EE / Oracle standards. Arkarta does not copy Java APIs. It keeps the proven enterprise boundaries and re-expresses them with Go-native design.

## 2. Go-Native Rules

- Use `context.Context` for cancellation, deadlines, trace propagation, and request scope.
- Use explicit `error` returns instead of Java exception hierarchies.
- Use small interfaces and function adapters instead of inheritance.
- Use explicit or generated registration instead of runtime classpath scanning.
- Preserve `net/http` interoperability.
- Keep standards package-scoped and independently testable.
- Back compatibility claims with TCKs.

## 3. Standard Layers

```text
goark.dev/arkarta/servlet              Servlet Core and container contracts
goark.dev/arkarta/servlet/*            Servlet profiles and adapters
goark.dev/arkarta/web                  MVC/REST composition layer
goark.dev/arkarta/json                 JSON codec standard and default sonic implementation
goark.dev/arkarta/json/sonic           Sonic compatibility package
goark.dev/arkarta/validation           Validation standard
goark.dev/arkarta/websocket            WebSocket standard
goark.dev/arkarta/websocket/frame      RFC 6455 frame layer
goark.dev/arkarta/websocket/servlet    Servlet Upgrade adapter for WebSocket
goark.dev/arkarta/security             Root enterprise security contracts
goark.dev/arkarta/*/tck                Executable compatibility tests
```

Concrete containers such as Goark Tomcat and Goark Jetty implement these contracts. They are not part of the Arkarta standard module.

## 4. Servlet Responsibility

`goark.dev/arkarta/servlet` defines the low-level application/container boundary:

- Request and response models.
- Handler and Servlet contracts.
- Filter chain and dispatcher types.
- WebApp context and lifecycle.
- Path mapping and dispatching.
- Error model and error pages.
- `net/http` interoperability.
- Optional profiles: Session, Multipart, Async/Stream, Upgrade, Native I/O, and Servlet declarative security.
- Servlet TCK entry points.

Servlet does not own controller method binding, JSON encoding, enterprise authentication providers, templates, or business routing syntax.

## 5. Web Responsibility

`goark.dev/arkarta/web` defines the Goark MVC/REST composition layer:

- Method router with path variables.
- Route groups and group-scoped interceptors.
- Automatic `HEAD` for `GET` and automatic `OPTIONS` with stable `Allow`.
- Query, path, header, cookie, form, and multipart binding helpers.
- Parameter conversion helpers.
- JSON binding and Validation integration.
- Unified `Result` model for JSON, text, and no-content responses.
- Response advice.
- Error mapping.
- Web TCK for container integrations.

Controller method discovery and generated registration belong to Goark tooling, not to runtime reflection in the standard.

## 6. JSON Responsibility

`goark.dev/arkarta/json` defines a codec contract inspired by JSON-B and JSON-P:

- `Codec`, `Encoder`, and `Decoder` interfaces.
- bytedance sonic as the only built-in implementation.
- Streaming encode/decode.
- Input size limit.
- Unknown-field rejection.
- Number precision mode.
- HTML escaping and indentation controls.

`goark.dev/arkarta/json/sonic` is the sonic compatibility entrypoint. `goark.dev/arkarta/json/tck` verifies sonic-compatible codec behavior.

## 7. Validation Responsibility

`goark.dev/arkarta/validation` defines Go-native validation:

- Struct-tag field constraints.
- Nested struct and slice validation.
- Validation groups through `arkarta-groups`.
- Message resolver extension point.
- Object-level constraints.
- Built-in constraints: `required`, `notblank`, `min`, `max`, `len`, `email`, `oneof`, `regexp`, `url`, `uuid`, `gt`, `gte`, `lt`, `lte`, `contains`, `startswith`, and `endswith`.
- Custom constraint registration.
- Aggregated validation result and stable validation error contract.

The standard intentionally avoids Java annotation scanning. Constraints are explicit Go values or struct tags parsed by the validator.

## 8. WebSocket Responsibility

`goark.dev/arkarta/websocket` defines the WebSocket standard:

- HTTP Upgrade handshake.
- Subprotocol negotiation.
- permessage-deflate extension negotiation.
- Endpoint and session contracts.
- Message and close status model.
- Connection SPI and service loop.
- JSON text codec.
- WebSocket TCK.

`websocket/frame` owns RFC 6455 frame encoding/decoding. `websocket/servlet` adapts Servlet Upgrade connections to WebSocket sessions.

## 9. Security Responsibility

Security is split intentionally:

- `servlet/security` defines Servlet declarative security: request-bound Principal, Basic authentication, Realm, role constraints, method constraints, run-as, transport guarantee, and security filter.
- Root `security` defines enterprise contracts independent of Servlet: Principal, Authority, Authentication, Credential, AuthenticationManager, SecurityContext, Authorizer, and authorization decisions.

Concrete identity providers, OAuth2/OIDC/JWT, CSRF, CORS, password storage, and policy engines belong to security implementations layered above this standard.

## 10. TCK Policy

A module or container may only claim Arkarta compatibility for the profiles it passes. Relevant TCKs include:

- `servlet/tck` for Servlet Core and profiles.
- `web/tck` for Web, JSON, and Validation integration over a container.
- `websocket/tck` for WebSocket handshake, endpoint, compression, and frame behavior.
- `json/tck` for JSON codec implementations.

Release gates for this repository:

```shell
go test ./...
go test -race ./...
go vet ./...
gofmt -l .
```

`gofmt -l .` must print no files.

## 11. Version Policy

`v0.0.1` freezes the first public preview of the standard. Before `v0.1.0`, source-incompatible corrections are still allowed when needed to fix standard quality. After `v0.1.0`, public APIs should evolve additively whenever possible.
