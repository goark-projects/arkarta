package container

// Metadata 描述容器实现的身份和能力。
type Metadata struct {
	name     string
	version  string
	profiles []Profile
	limits   map[string]string
}

// NewMetadata 创建容器元数据。
func NewMetadata(name, version string, profiles []Profile, limits map[string]string) Metadata {
	return Metadata{
		name:     name,
		version:  version,
		profiles: cloneProfiles(profiles),
		limits:   cloneStringMap(limits),
	}
}

// Name 返回容器名称。
func (m Metadata) Name() string {
	return m.name
}

// Version 返回容器版本。
func (m Metadata) Version() string {
	return m.version
}

// Profiles 返回容器支持的 Profile 副本。
func (m Metadata) Profiles() []Profile {
	return cloneProfiles(m.profiles)
}

// Limits 返回容器限制参数副本。
func (m Metadata) Limits() map[string]string {
	return cloneStringMap(m.limits)
}

// Supports 判断容器是否支持指定 Profile。
func (m Metadata) Supports(profile Profile) bool {
	return SupportsProfile(m.profiles, profile)
}
