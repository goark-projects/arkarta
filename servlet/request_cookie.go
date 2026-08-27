package servlet

import "net/http"

// Cookies 返回请求 Cookie 的防御性副本。
func (r *Request) Cookies() []*http.Cookie {
	cookies := r.httpRequest.Cookies()
	if len(cookies) == 0 {
		return nil
	}
	result := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		cloned := *cookie
		result = append(result, &cloned)
	}
	return result
}
