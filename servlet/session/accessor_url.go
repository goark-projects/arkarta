package session

import "goark.dev/arkarta/servlet"

// EncodeURL 使用当前请求会话编码普通 URL。
func (a *Accessor) EncodeURL(req *servlet.Request, rawURL string) (string, error) {
	return a.encodeURL(req, rawURL)
}

// EncodeRedirectURL 使用当前请求会话编码重定向 URL。
func (a *Accessor) EncodeRedirectURL(req *servlet.Request, rawURL string) (string, error) {
	return a.encodeURL(req, rawURL)
}

func (a *Accessor) encodeURL(req *servlet.Request, rawURL string) (string, error) {
	if a == nil || a.manager == nil {
		return "", ErrNilManager
	}
	if req == nil {
		return "", ErrNilRequest
	}
	if !a.tracking.Allows(TrackingURL) {
		return rawURL, nil
	}
	id, ok := rewriteSessionID(a, req)
	if !ok {
		return rawURL, nil
	}
	rewriter, err := NewURLRewriter(
		WithRewriteCookieName(a.cookie.name),
		WithCookiePreferred(a.tracking.Allows(TrackingCookie)),
	)
	if err != nil {
		return "", err
	}
	return rewriter.EncodeURL(req, rawURL, id)
}

func rewriteSessionID(accessor *Accessor, req *servlet.Request) (string, bool) {
	if current, ok := Current(req); ok {
		return current.ID(), true
	}
	return accessor.RequestedID(req)
}
