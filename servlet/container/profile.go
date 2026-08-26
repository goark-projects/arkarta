package container

// Profile 表示容器通过的 Servlet 能力集合。
type Profile string

const (
	// ProfileCore 是所有兼容容器必须实现的核心能力。
	ProfileCore Profile = "core"
	// ProfileSession 表示容器支持会话能力。
	ProfileSession Profile = "session"
	// ProfileMultipart 表示容器支持 multipart/form-data。
	ProfileMultipart Profile = "multipart"
	// ProfileAsyncStream 表示容器支持异步和流式响应。
	ProfileAsyncStream Profile = "async-stream"
	// ProfileUpgrade 表示容器支持协议升级。
	ProfileUpgrade Profile = "upgrade"
	// ProfileNativeIO 表示容器暴露原生 I/O 能力。
	ProfileNativeIO Profile = "native-io"
)

// SupportsProfile 判断 Profile 列表是否包含指定能力。
func SupportsProfile(profiles []Profile, profile Profile) bool {
	for _, item := range profiles {
		if item == profile {
			return true
		}
	}
	return false
}
