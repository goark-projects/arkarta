package session

import (
	"context"
	"strings"

	"goark.dev/arkarta/servlet"
)

const (
	// AttributeCurrentSession 保存当前请求关联的 Session。
	AttributeCurrentSession = "arkarta.servlet.session.current"
	// AttributeRequestedSessionID 保存客户端提交的原始 Session ID。
	AttributeRequestedSessionID = "arkarta.servlet.session.requested_id"
	// AttributeRequestedSessionIDSource 保存客户端提交 Session ID 的来源。
	AttributeRequestedSessionIDSource = "arkarta.servlet.session.requested_id_source"
)

// Accessor 在 Request/Response 边界实现 Servlet Session 绑定语义。
type Accessor struct {
	manager  Manager
	cookie   CookieConfig
	tracking TrackingPolicy
}

// NewAccessor 创建请求会话访问器。
func NewAccessor(manager Manager, options ...AccessorOption) (*Accessor, error) {
	if manager == nil {
		return nil, ErrNilManager
	}
	accessor := &Accessor{
		manager:  manager,
		cookie:   defaultCookieConfig(),
		tracking: DefaultTrackingPolicy(),
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(accessor); err != nil {
			return nil, err
		}
	}
	return accessor, nil
}

// NewAccessorForWebApp 创建使用 WebApp 级 Session 配置的访问器。
func NewAccessorForWebApp(manager Manager, app *servlet.WebApp, options ...AccessorOption) (*Accessor, error) {
	if config, ok := CookieConfigFor(app); ok {
		options = append([]AccessorOption{WithCookieConfig(config)}, options...)
	}
	return NewAccessor(manager, options...)
}

// Manager 返回底层 Session Manager。
func (a *Accessor) Manager() Manager {
	if a == nil {
		return nil
	}
	return a.manager
}

// CookieConfig 返回 Cookie 配置副本。
func (a *Accessor) CookieConfig() CookieConfig {
	if a == nil {
		return CookieConfig{}
	}
	return a.cookie
}

// TrackingPolicy 返回会话跟踪策略。
func (a *Accessor) TrackingPolicy() TrackingPolicy {
	if a == nil {
		return DefaultTrackingPolicy()
	}
	return a.tracking
}

// Get 返回当前请求 Session；create 为 true 时会创建并写回 Cookie。
func (a *Accessor) Get(ctx context.Context, req *servlet.Request, res servlet.Response, create bool) (Session, bool, error) {
	if a == nil || a.manager == nil {
		return nil, false, ErrNilManager
	}
	if req == nil {
		return nil, false, ErrNilRequest
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if current, ok := Current(req); ok {
		return current, true, nil
	}
	if requestedID, ok := a.RequestedID(req); ok {
		current, found, err := a.manager.Get(ctx, requestedID)
		if err != nil {
			return nil, false, err
		}
		if found {
			req.SetAttribute(AttributeCurrentSession, current)
			return current, true, nil
		}
	}
	if !create {
		return nil, false, nil
	}
	if a.tracking.Allows(TrackingCookie) && res == nil {
		return nil, false, servlet.ErrNilResponse
	}
	created, err := a.createSession(ctx, req)
	if err != nil {
		return nil, false, err
	}
	req.SetAttribute(AttributeCurrentSession, created)
	if a.tracking.Allows(TrackingCookie) && res != nil {
		if err := a.writeCookie(req, res, created.ID()); err != nil {
			_ = created.Invalidate()
			req.SetAttribute(AttributeCurrentSession, nil)
			return nil, false, err
		}
	}
	return created, true, nil
}

// RequestedID 返回客户端提交的原始 Session ID。
func (a *Accessor) RequestedID(req *servlet.Request) (string, bool) {
	id, _, ok := a.requestedID(req)
	return id, ok
}

// RequestedIDSource 返回客户端提交 Session ID 的来源。
func (a *Accessor) RequestedIDSource(req *servlet.Request) (TrackingMode, bool) {
	_, source, ok := a.requestedID(req)
	return source, ok
}

func (a *Accessor) requestedID(req *servlet.Request) (string, TrackingMode, bool) {
	if a == nil || req == nil {
		return "", "", false
	}
	if value, ok := req.Attribute(AttributeRequestedSessionID); ok {
		id, _ := value.(string)
		sourceValue, _ := req.Attribute(AttributeRequestedSessionIDSource)
		source, _ := sourceValue.(TrackingMode)
		return id, source, id != ""
	}
	if a.tracking.Allows(TrackingCookie) {
		if cookie, err := req.Cookie(a.cookie.name); err == nil && cookie.Value != "" {
			req.SetAttribute(AttributeRequestedSessionID, cookie.Value)
			req.SetAttribute(AttributeRequestedSessionIDSource, TrackingCookie)
			return cookie.Value, TrackingCookie, true
		}
	}
	if a.tracking.Allows(TrackingURL) {
		if id, ok := pathSessionID(req.Path(), DefaultURLParameterName); ok {
			req.SetAttribute(AttributeRequestedSessionID, id)
			req.SetAttribute(AttributeRequestedSessionIDSource, TrackingURL)
			return id, TrackingURL, true
		}
	}
	if a.tracking.Allows(TrackingSSL) && req.IsSecure() {
		if id := req.ConnectionInfo().ID(); id != "" {
			req.SetAttribute(AttributeRequestedSessionID, id)
			req.SetAttribute(AttributeRequestedSessionIDSource, TrackingSSL)
			return id, TrackingSSL, true
		}
	}
	req.SetAttribute(AttributeRequestedSessionID, "")
	req.SetAttribute(AttributeRequestedSessionIDSource, nil)
	return "", "", false
}

func (a *Accessor) createSession(ctx context.Context, req *servlet.Request) (Session, error) {
	if !a.tracking.Allows(TrackingCookie) && !a.tracking.Allows(TrackingURL) && a.tracking.Allows(TrackingSSL) {
		if id, source, ok := a.requestedID(req); ok && source == TrackingSSL {
			if manager, supports := a.manager.(IDBoundManager); supports {
				return manager.CreateWithID(ctx, id)
			}
		}
	}
	return a.manager.Create(ctx)
}

func pathSessionID(path, name string) (string, bool) {
	prefix := name + "="
	for _, segment := range splitPathSegments(path) {
		for _, part := range splitPathParameters(segment) {
			if value, ok := strings.CutPrefix(part, prefix); ok && value != "" {
				return value, true
			}
		}
	}
	return "", false
}

func splitPathSegments(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func splitPathParameters(segment string) []string {
	parts := strings.Split(segment, ";")
	if len(parts) <= 1 {
		return nil
	}
	return parts[1:]
}

func (a *Accessor) writeCookie(req *servlet.Request, res servlet.Response, id string) error {
	if !a.tracking.Allows(TrackingCookie) {
		return nil
	}
	return servlet.AddCookie(res, a.cookie.cookie(req, id))
}

// RequestedIDValid 判断客户端提交的 Session ID 是否仍然有效。
func (a *Accessor) RequestedIDValid(ctx context.Context, req *servlet.Request) (bool, error) {
	if a == nil || a.manager == nil {
		return false, ErrNilManager
	}
	if req == nil {
		return false, ErrNilRequest
	}
	id, ok := a.RequestedID(req)
	if !ok {
		return false, nil
	}
	_, found, err := a.manager.Get(ctx, id)
	return found, err
}

// ChangeID 轮换当前请求关联的 Session ID 并写回 Cookie。
func (a *Accessor) ChangeID(ctx context.Context, req *servlet.Request, res servlet.Response) (string, error) {
	if a == nil || a.manager == nil {
		return "", ErrNilManager
	}
	if req == nil {
		return "", ErrNilRequest
	}
	if a.tracking.Allows(TrackingCookie) && res == nil {
		return "", servlet.ErrNilResponse
	}
	current, ok := Current(req)
	if !ok {
		var err error
		current, ok, err = a.Get(ctx, req, nil, false)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", ErrSessionNotFound
		}
	}
	renewed, err := a.manager.RenewID(ctx, current.ID())
	if err != nil {
		return "", err
	}
	req.SetAttribute(AttributeCurrentSession, renewed)
	if err := a.writeCookie(req, res, renewed.ID()); err != nil {
		return "", err
	}
	return renewed.ID(), nil
}
