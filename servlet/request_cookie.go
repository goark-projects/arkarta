package servlet

// Cookies 返回请求 Cookie 的防御性副本。
func (r *Request) Cookies() []*Cookie {
	r.cookiesOnce.Do(func() {
		r.cookies = parseRequestCookies(r.header)
	})
	cookies := r.cookies
	if len(cookies) == 0 {
		return nil
	}
	result := make([]*Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		cloned := *cookie
		result = append(result, &cloned)
	}
	return result
}

// Cookie 返回指定名称的 Cookie。
func (r *Request) Cookie(name string) (*Cookie, error) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == name {
			return cookie, nil
		}
	}
	return nil, ErrNoCookie
}
