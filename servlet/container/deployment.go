package container

import (
	"errors"

	"goark.dev/arkarta/servlet"
)

// ErrNilWebApp 表示部署描述缺少 WebApp。
var ErrNilWebApp = errors.New("arkarta/servlet/container: web app is nil")

// ErrEmptyDeployment 表示部署描述没有任何路径映射。
var ErrEmptyDeployment = errors.New("arkarta/servlet/container: deployment has no mapping")

// Deployment 描述一个待部署的 Web 应用。
type Deployment struct {
	webApp   *servlet.WebApp
	mappings []Mapping
	profiles []Profile
}

// DeploymentOption 定制部署描述。
type DeploymentOption func(*Deployment) error

// WithMapping 添加路径映射。
func WithMapping(pattern string, handler servlet.Handler, filters ...servlet.Filter) DeploymentOption {
	return func(deployment *Deployment) error {
		mapping, err := NewMapping(pattern, handler, filters...)
		if err != nil {
			return err
		}
		deployment.mappings = append(deployment.mappings, mapping)
		return nil
	}
}

// WithServlet 添加带生命周期的 Servlet 路径映射。
func WithServlet(pattern, name string, handler servlet.Servlet, filters ...servlet.Filter) DeploymentOption {
	return func(deployment *Deployment) error {
		mapping, err := newServletMapping(pattern, name, handler, filters...)
		if err != nil {
			return err
		}
		deployment.mappings = append(deployment.mappings, mapping)
		return nil
	}
}

// WithProfile 声明部署需要的 Profile。
func WithProfile(profile Profile) DeploymentOption {
	return func(deployment *Deployment) error {
		if profile == "" || SupportsProfile(deployment.profiles, profile) {
			return nil
		}
		deployment.profiles = append(deployment.profiles, profile)
		return nil
	}
}

// NewDeployment 创建部署描述。
func NewDeployment(webApp *servlet.WebApp, options ...DeploymentOption) (*Deployment, error) {
	if webApp == nil {
		return nil, ErrNilWebApp
	}
	deployment := &Deployment{
		webApp:   webApp,
		profiles: []Profile{ProfileCore},
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(deployment); err != nil {
			return nil, err
		}
	}
	return deployment, nil
}

// WebApp 返回部署所属应用上下文。
func (d *Deployment) WebApp() *servlet.WebApp {
	return d.webApp
}

// Mappings 返回路径映射副本。
func (d *Deployment) Mappings() []Mapping {
	return cloneMappings(d.mappings)
}

// Profiles 返回部署需要的 Profile 副本。
func (d *Deployment) Profiles() []Profile {
	return cloneProfiles(d.profiles)
}

// Handler 将部署描述构造成可执行处理器。
func (d *Deployment) Handler() (servlet.Handler, error) {
	if len(d.mappings) == 0 {
		return nil, ErrEmptyDeployment
	}
	router := servlet.NewRouter()
	dispatchers := servlet.NewDispatcherRegistry(router)
	for _, mapping := range d.mappings {
		if err := router.Handle(mapping.Pattern(), mapping.servletHandler()); err != nil {
			return nil, err
		}
		if path := dispatcherPathForPattern(mapping.Pattern()); path != "" {
			dispatchers.RegisterName(mapping.Name(), path)
		}
	}
	d.webApp.SetDispatcherProvider(dispatchers)
	return router, nil
}

func dispatcherPathForPattern(pattern string) string {
	if pattern == "" {
		return ""
	}
	if pattern == "/" {
		return "/"
	}
	if len(pattern) > 2 && pattern[len(pattern)-2:] == "/*" {
		base := pattern[:len(pattern)-2]
		if base == "" {
			return "/"
		}
		return base
	}
	if pattern[0] == '/' {
		return pattern
	}
	return ""
}
