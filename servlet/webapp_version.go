package servlet

const (
	// ServletSpecMajorVersion 是 Arkarta Servlet 当前对齐的 Jakarta Servlet 主版本。
	ServletSpecMajorVersion = 6
	// ServletSpecMinorVersion 是 Arkarta Servlet 当前对齐的 Jakarta Servlet 次版本。
	ServletSpecMinorVersion = 1
	// ArkartaServletMajorVersion 是 Arkarta Servlet 标准主版本。
	ArkartaServletMajorVersion = 1
	// ArkartaServletMinorVersion 是 Arkarta Servlet 标准次版本。
	ArkartaServletMinorVersion = 0
)

// ServletContext 是 WebApp 的 Servlet 语义别名。
type ServletContext = WebApp

// EffectiveMajorVersion 返回当前上下文采用的 Servlet 主版本。
func (a *WebApp) EffectiveMajorVersion() int {
	return ServletSpecMajorVersion
}

// EffectiveMinorVersion 返回当前上下文采用的 Servlet 次版本。
func (a *WebApp) EffectiveMinorVersion() int {
	return ServletSpecMinorVersion
}

// ArkartaMajorVersion 返回 Arkarta Servlet 标准主版本。
func (a *WebApp) ArkartaMajorVersion() int {
	return ArkartaServletMajorVersion
}

// ArkartaMinorVersion 返回 Arkarta Servlet 标准次版本。
func (a *WebApp) ArkartaMinorVersion() int {
	return ArkartaServletMinorVersion
}
