# Arkarta Servlet 1.0

Language: English | [简体中文](arkarta-servlet-1.0.zh-CN.md)

Status: Release Candidate 1

Module: `goark.dev/arkarta/servlet`

Target release: `v0.0.1`

Date: 2026-08-27

## 1. Purpose

Arkarta Servlet defines the stable boundary between Goark Web applications and Web containers. It uses Jakarta Servlet 6.1 as the primary reference standard and Java Servlet / Oracle Java EE history as background, but it does not claim Jakarta compatibility and does not reuse `jakarta.servlet` or `javax.servlet` namespaces.

An Arkarta container is compatible only when it implements the relevant contracts and passes the relevant Arkarta TCKs.

## 2. Normative Language

- Must: required for compatibility.
- Should: expected unless a clear engineering reason exists.
- May: optional behavior.
- Container: a runtime that deploys applications, owns lifecycle, dispatches requests, and exposes supported profiles.
- Application: user code registered into a container.

## 3. Design Principles

- Go first: use `context.Context`, `error`, interfaces, functions, and `net/http`.
- Container neutral: no assumption about Tomcat, Jetty, `net/http`, epoll, kqueue, io_uring, or any event loop.
- Small core, explicit profiles: Session, Multipart, Async, Upgrade, Native I/O, Security, and WebSocket integration remain separate packages.
- No Java inheritance model: no `GenericServlet`, `HttpServlet`, annotation scanning, WAR/EAR model, or classloader semantics.
- Compatibility is executable: required behavior must be covered by TCK.

## 4. Core Profile

Every Arkarta Servlet container must support:

- `Handler`, `HandlerFunc`, and lifecycle-aware `Servlet`.
- `Request` and `Response`.
- `Filter`, `ManagedFilter`, and `Chain`.
- `WebApp` context, init parameters, attributes, resources, listeners, MIME mapping, temp directory, session timeout, and version metadata.
- Servlet path mapping: exact, longest-prefix, extension, and default mappings.
- Dispatch types: request, forward, include, error, and async.
- RequestDispatcher forward/include/error behavior.
- Status errors, panic recovery, error pages, and committed-response safety.
- `net/http` interoperability.

## 5. Request Contract

`Request` must expose:

- Method, protocol, scheme, host, remote address, secure flag, and connection ID.
- `RequestURI`, `RequestURL`, `QueryString`, `ContextPath`, `Path`, `ServletPath`, `PathInfo`, and `RequestMapping`.
- Query values and combined Servlet parameters from query plus `application/x-www-form-urlencoded` body.
- Stable sorted parameter names.
- Headers, typed header helpers, cookies, and cookie snapshots.
- Locale parsing from `Accept-Language`.
- Accept media type parsing and content negotiation helpers.
- Trailer fields and raw `*http.Request` interoperability.
- Body reader and request attributes.

Query values must remain query-only. Servlet parameters are the combined parameter view.

## 6. Response Contract

`Response` must expose:

- Headers, status, body writes, string writes, flush, committed flag, reset, and body writer.
- Cookie helpers.
- Redirect and error helpers.
- Content-Type, charset, content length, typed header, locale, and trailer helpers.

The first body write commits the response. `Reset` must fail after commit. Error handling must not leak stack traces or local paths by default.

## 7. Filter and Dispatch Contract

Containers must provide deterministic filter order. Filters may short-circuit by not calling `chain.Next`, but must not call `Next` more than once. Filter bindings must respect dispatcher types and URL-pattern matching.

Forward must happen before response commit. Include must not overwrite the main response status. Error dispatch must expose stable request attributes for status, exception, request URI, servlet name, and message.

## 8. Optional Profiles Implemented in v0.0.1

### Session

`servlet/session` defines session manager and accessor contracts, cookie and URL tracking, SSL tracking policy, ID rotation, requested ID validation, URL rewriting, store SPI, memory store, passivation/activation, and listeners.

### Multipart

`servlet/multipart` defines parser options, max body and file size, memory threshold, temp directory, submitted filename normalization, `Part`, request binding, and cleanup.

### Async and Stream

`servlet/async` defines explicit async lifecycle, completion, timeout, errors, dispatch counts, await, and stream writes. Go goroutines alone are not treated as Servlet async lifecycle; applications must opt into the profile.

### Upgrade

`servlet/upgrade` defines connection ownership transfer and `net/http` hijack adaptation. WebSocket protocol behavior is owned by `goark.dev/arkarta/websocket`.

### Native I/O

`servlet/nativeio` defines file region sending, capability declaration, send strategy reporting, portable fallback sender, invalid region handling, and cancellation behavior.

### Declarative Security

`servlet/security` defines request-bound Principal, Basic authenticator, Realm, role constraints, method constraints, role mapping, run-as, transport guarantee, and security filter.

## 9. Container SPI

`servlet/container` defines:

- `Container`
- `Application`
- `Deployment`
- registration snapshot conversion
- application lifecycle
- profile declarations
- metadata for container name, version, supported profiles, and limits

Containers must validate mapping conflicts, registration freeze rules, lifecycle ordering, and profile dependencies during deployment.

## 10. Servlet 6.1 Coverage Matrix

| Servlet area | Arkarta v0.0.1 status |
| --- | --- |
| Request path and mapping details | Implemented |
| Query/form parameters and parameter names | Implemented |
| Header, locale, cookie, trailer, and connection metadata | Implemented |
| Accept negotiation and response Content-Type helpers | Implemented |
| Response status, headers, body, cookies, redirects, errors, charset, trailers | Implemented |
| Servlet path mapping | Implemented |
| RequestDispatcher forward/include/error | Implemented |
| Dynamic Servlet, Filter, and Listener registration | Implemented |
| Filter lifecycle and dispatcher type filtering | Implemented |
| WebApp context and listeners | Implemented |
| Session profile | Implemented |
| Multipart profile | Implemented |
| Static resources and welcome files | Implemented |
| Async and stream profile | Implemented |
| Upgrade profile | Implemented |
| WebSocket adapter | Implemented through `websocket/servlet` |
| Native I/O profile | Implemented |
| Declarative security profile | Implemented |
| Java annotation scanning and WAR/EAR packaging | Not included |
| JSP, JSTL, EL, Java classloader semantics | Not included |
| HTTP/2 server push requirement | Not included |

## 11. Compatibility

A container must run the TCK entry points that match its claimed profiles:

- `servlet/tck.RunCoreHTTP`
- `servlet/tck.RunLifecycle`
- `servlet/tck.RunErrorPages`
- `servlet/tck.RunSessionManager`
- `servlet/tck.RunSessionRequestBinding`
- `servlet/tck.RunMultipartParser`
- `servlet/tck.RunAsyncLifecycle`
- `servlet/tck.RunSecurity`
- `servlet/tck.RunNativeIO`
- `servlet/tck.RunStaticResources`
- `servlet/tck.RunHTTPContainer`

Passing Core does not imply passing optional profiles.

## 12. Versioning

Arkarta Servlet 1.x should evolve additively after the preview period. Optional capabilities should use new packages, small interfaces, or options instead of changing released method signatures.
