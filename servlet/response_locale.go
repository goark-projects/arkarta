package servlet

// SetLocale 设置响应 Content-Language。
func SetLocale(res Response, locale Locale) error {
	if locale.Tag() == "" {
		return SetHeader(res, "Content-Language", "")
	}
	return SetHeader(res, "Content-Language", locale.Tag())
}

// ResponseLocale 返回响应 Content-Language。
func ResponseLocale(res Response) (Locale, bool) {
	value, ok := HeaderValue(res, "Content-Language")
	if !ok {
		return Locale{}, false
	}
	return NewLocale(value)
}
