package session

import (
	"context"

	"goark.dev/arkarta/servlet"
)

const (
	// AttributeCurrentSession 保存当前请求关联的 Session。
	AttributeCurrentSession = "arkarta.servlet.session.current"
	// AttributeRequestedSessionID 保存客户端提交的原始 Session ID。
	AttributeRequestedSessionID = "arkarta.servlet.session.requested_id"
)

// Accessor 在 Request/Response 边界实现 Servlet Session 绑定语义。
type Accessor struct {
	manager Manager
	cookie  CookieConfig
}

// NewAccessor 创建请求会话访问器。
func NewAccessor(manager Manager, options ...AccessorOption) (*Accessor, error) {
	if manager == nil {
		return nil, ErrNilManager
	}
	accessor := &Accessor{
		manager: manager,
		cookie:  defaultCookieConfig(),
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
	if res == nil {
		return nil, false, servlet.ErrNilResponse
	}
	created, err := a.manager.Create(ctx)
	if err != nil {
		return nil, false, err
	}
	req.SetAttribute(AttributeCurrentSession, created)
	if err := a.writeCookie(req, res, created.ID()); err != nil {
		_ = created.Invalidate()
		req.SetAttribute(AttributeCurrentSession, nil)
		return nil, false, err
	}
	return created, true, nil
}

// RequestedID 返回客户端提交的原始 Session ID。
func (a *Accessor) RequestedID(req *servlet.Request) (string, bool) {
	if a == nil || req == nil {
		return "", false
	}
	if value, ok := req.Attribute(AttributeRequestedSessionID); ok {
		id, _ := value.(string)
		return id, id != ""
	}
	cookie, err := req.Cookie(a.cookie.name)
	if err != nil || cookie.Value == "" {
		req.SetAttribute(AttributeRequestedSessionID, "")
		return "", false
	}
	req.SetAttribute(AttributeRequestedSessionID, cookie.Value)
	return cookie.Value, true
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
	if res == nil {
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

func (a *Accessor) writeCookie(req *servlet.Request, res servlet.Response, id string) error {
	return servlet.AddCookie(res, a.cookie.cookie(req, id))
}
