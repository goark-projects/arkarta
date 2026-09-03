# Changelog

Language: English | [简体中文](CHANGELOG.zh-CN.md)

This project follows Go module semantic versioning. `v0.0.x` releases are early previews: public APIs are designed as standard contracts, but necessary source-incompatible corrections may still happen before `v0.1.0`.

Release notes: [v0.0.2](docs/releases/v0.0.2.md) | [v0.0.1](docs/releases/v0.0.1.md)

## [Unreleased]

## [0.0.2] - 2026-09-03

### Added

- Added transport-neutral Servlet `Header`, `Cookie`, and `RequestInput` contracts for container implementations.
- Added request mutation methods required by forwarding and dispatching without exposing a native HTTP request.
- Added transport-neutral TCK drivers and a `net/http` reference driver.
- Added the Upgrade container profile and transport-neutral WebSocket handshake entry point.
- Added Web validation groups.
- Added a configurable URL-encoded form body limit with a safe 10 MiB default.

### Changed

- Decoupled Servlet request, response, multipart, Basic authentication, cookies, headers, and WebSocket integration from `net/http` concrete types.
- Removed the built-in `encoding/json` codec and made bytedance sonic the single Arkarta JSON implementation.
- Made the default servlet lifecycle-aware and required async worker quiescence during completion.
- Passed request options through the `net/http` reference adapter and avoided repeated context-path derivation on the request hot path.

### Fixed

- Fixed URL-encoded form parsing for transport-neutral request bodies.
- Preserved request and response behavior across both the `net/http` reference path and native container bridges.

### Compatibility

- `Request.Header()` and `Response.Header()` now return `servlet.Header` instead of `http.Header`.
- Request cookies now use `servlet.Cookie` and missing cookies return `servlet.ErrNoCookie`.
- `json.StandardCodec` and `json.NewStandardCodec` were removed; use `json.NewCodec` or `json.NewSonicCodec`.
- These source-incompatible corrections are permitted while Arkarta remains in the `v0.0.x` preview series.

## [0.0.1] - 2026-08-27

### Added

- Initialized the `goark.dev/arkarta` module as the Goark enterprise Web standard collection.
- Added Servlet Core APIs: `Handler`, `Servlet`, `Request`, `Response`, `Filter`, `Chain`, `WebApp`, lifecycle contracts, request dispatching, and structured error semantics.
- Added Servlet 6.1-aligned request/response capabilities: path details, parameters, parameter names, cookies, mapping elements, headers, locale, trailers, connection metadata, content negotiation, response cookies, redirect, error, charset, content length, and typed header helpers.
- Added `servlet/container` SPI for deployments, applications, registration snapshots, lifecycle, profile declarations, and container metadata.
- Added `servlet/registration` dynamic registration metadata for Servlet, Filter, and Listener definitions, including mappings, init params, multipart config, security constraints, listener ordering, and frozen snapshots.
- Added `servlet/nethttp` reference adapter and minimal reference container based on `net/http`.
- Added `servlet/resource` static resource provider, `fs.FS` provider, default servlet, conditional requests, weak ETag handling, multi-range responses, and welcome file resolution.
- Added `servlet/session` session profile with manager SPI, memory implementation, cookie and URL tracking, SSL tracking policy, requested ID validation, ID rotation, URL rewriting, store SPI, passivation/activation, and listeners.
- Added `servlet/multipart` multipart profile with parser, part API, request binding, size limits, memory threshold, temp directory handling, submitted filename normalization, and cleanup semantics.
- Added `servlet/async` async and stream profile with explicit completion, await, timeout, error events, dispatch counts, async dispatch, and serialized stream writes.
- Added `servlet/upgrade` upgrade profile with connection handoff contracts and `net/http` hijack adapter.
- Added `servlet/nativeio` native I/O profile with file region sender contract, capability declaration, send strategy reporting, portable fallback implementation, and TCK.
- Added `servlet/security` declarative security profile with Principal, Basic authentication, Realm, role constraints, method constraints, run-as, transport guarantee, and security filter.
- Added `servlet/tck` compatibility tests for Core HTTP, lifecycle, dispatching, error pages, registration, WebApp context, resources, session, multipart, async, security, native I/O, and HTTP container entry points.
- Added `websocket` standard package for RFC 6455 handshake, subprotocol negotiation, permessage-deflate negotiation, endpoints, sessions, messages, close status, connection SPI, service loop, and JSON text codec.
- Added `websocket/frame` RFC 6455 frame layer with masking, extended lengths, control-frame validation, close payload handling, and fragmentation assembly.
- Added `websocket/servlet` integration for Servlet Upgrade, HTTP 101 response writing, frame connection adaptation, and endpoint service helper.
- Added `websocket/tck` compatibility tests for handshake, endpoint lifecycle, compression, and frame codec.
- Added `json` standard package with `encoding/json` default codec, streaming encoder/decoder, max input size, unknown-field gate, number precision mode, and package helpers.
- Added `json/sonic` high-performance codec implementation based on `github.com/bytedance/sonic`.
- Added `json/tck` codec compatibility tests shared by the standard and sonic implementations.
- Added `validation` standard package with struct-tag constraints, nested validation, validation groups, message resolver, object constraints, built-in constraints, custom constraints, and aggregated validation errors.
- Added `web` MVC/REST composition layer with method routing, route groups, automatic HEAD/OPTIONS, path/query/form/multipart binding, parameter conversion helpers, unified results, response advice, error mapping, and interceptors.
- Added `web/tck` compatibility tests for Web routing, JSON binding, Validation mapping, content negotiation, automatic method semantics, and form binding.
- Added root `security` contracts for Principal, Authority, Authentication, Credential, AuthenticationManager, SecurityContext, Authorizer, and authorization decisions.
- Added English default documentation and Simplified Chinese mirror documentation.

### Not Included

- Concrete Goark Tomcat, Goark Jetty, or other production Web containers.
- Java namespace compatibility with `jakarta.servlet` or `javax.servlet`.
- JSP, JSTL, Expression Language, WAR/EAR packaging, Java annotations, runtime classpath scanning, or Java-style inheritance APIs.
