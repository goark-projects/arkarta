package servlet

import "net/http"

func requestPath(httpRequest *http.Request) string {
	if httpRequest.URL == nil {
		return ""
	}
	return httpRequest.URL.Path
}

func requestURI(httpRequest *http.Request) string {
	if httpRequest.URL == nil {
		return ""
	}
	if uri := httpRequest.URL.EscapedPath(); uri != "" {
		return uri
	}
	if httpRequest.URL.Path != "" {
		return httpRequest.URL.Path
	}
	return "/"
}

func requestQueryString(httpRequest *http.Request) string {
	if httpRequest.URL == nil {
		return ""
	}
	return httpRequest.URL.RawQuery
}

func requestInputFromHTTP(httpRequest *http.Request) RequestInput {
	scheme := "http"
	if httpRequest.URL != nil && httpRequest.URL.Scheme != "" {
		scheme = httpRequest.URL.Scheme
	} else if httpRequest.TLS != nil {
		scheme = "https"
	}
	localAddr := ""
	if addr, ok := httpRequest.Context().Value(http.LocalAddrContextKey).(interface{ String() string }); ok && addr != nil {
		localAddr = addr.String()
	}
	return RequestInput{
		Context:       httpRequest.Context(),
		Method:        httpRequest.Method,
		Protocol:      httpRequest.Proto,
		Scheme:        scheme,
		Host:          httpRequest.Host,
		RequestURI:    requestURI(httpRequest),
		Path:          requestPath(httpRequest),
		QueryString:   requestQueryString(httpRequest),
		Header:        mapHeader(httpRequest.Header),
		Body:          httpRequest.Body,
		ContentLength: httpRequest.ContentLength,
		RemoteAddr:    httpRequest.RemoteAddr,
		LocalAddr:     localAddr,
		Trailer:       mapHeader(httpRequest.Trailer),
		TrailerReady: func() bool {
			for _, values := range httpRequest.Trailer {
				if values == nil {
					return false
				}
			}
			return true
		},
	}
}
