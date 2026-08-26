package container

import (
	"errors"
	"fmt"

	"goark.dev/arkarta/servlet"
	"goark.dev/arkarta/servlet/registration"
	"goark.dev/arkarta/servlet/security"
)

// ErrUnknownServletName 表示 Filter 引用了不存在的 Servlet 名称。
var ErrUnknownServletName = errors.New("arkarta/servlet/container: unknown servlet name")

// DeploymentFromRegistration 将注册快照转换为标准部署描述。
func DeploymentFromRegistration(app *servlet.WebApp, snapshot registration.Snapshot, options ...DeploymentOption) (*Deployment, error) {
	all := make([]DeploymentOption, 0, len(options)+1)
	all = append(all, WithRegistration(snapshot))
	all = append(all, options...)
	return NewDeployment(app, all...)
}

// WithRegistration 将动态注册快照追加到部署描述。
func WithRegistration(snapshot registration.Snapshot) DeploymentOption {
	return func(deployment *Deployment) error {
		if !snapshot.Frozen() {
			return registration.ErrSnapshotNotFrozen
		}
		if err := attachRegistrationListeners(deployment.webApp, snapshot.Listeners()); err != nil {
			return err
		}
		start := len(deployment.mappings)
		nameIndex := make(map[string][]int)
		for index, mapping := range deployment.mappings {
			if mapping.Name() != "" {
				nameIndex[mapping.Name()] = append(nameIndex[mapping.Name()], index)
			}
		}
		for _, descriptor := range snapshot.Servlets() {
			for _, pattern := range descriptor.Mappings() {
				loadOrder, hasLoadOrder := descriptor.LoadOnStartup()
				mapping, err := newRegistrationMapping(pattern, descriptor.Name(), descriptor.Handler(), descriptor.InitParams(), loadOrder, hasLoadOrder, descriptor.RunAsRole())
				if err != nil {
					return err
				}
				if err := attachSecurityFilter(&mapping, descriptor); err != nil {
					return err
				}
				deployment.mappings = append(deployment.mappings, mapping)
				nameIndex[descriptor.Name()] = append(nameIndex[descriptor.Name()], len(deployment.mappings)-1)
			}
		}
		return attachRegistrationFilters(deployment, start, nameIndex, snapshot.Filters())
	}
}

func attachSecurityFilter(mapping *Mapping, descriptor registration.ServletDescriptor) error {
	config, ok := descriptor.SecurityConfig()
	if !ok {
		return nil
	}
	dispatchers, err := servlet.NewDispatchTypes(servlet.DispatchRequest, servlet.DispatchForward, servlet.DispatchAsync)
	if err != nil {
		return err
	}
	binding, err := servlet.NewFilterBinding(
		descriptor.Name()+"#security",
		security.NewFilter(config),
		servlet.WithFilterDispatchTypes(dispatchers),
		servlet.WithFilterURLPattern(mapping.Pattern()),
	)
	if err != nil {
		return err
	}
	mapping.filterBindings = append([]servlet.FilterBinding{binding}, mapping.filterBindings...)
	return nil
}

func attachRegistrationListeners(app *servlet.WebApp, descriptors []registration.ListenerDescriptor) error {
	for _, descriptor := range descriptors {
		switch descriptor.Kind() {
		case registration.ListenerContext:
			listener, ok := descriptor.Listener().(servlet.ContextListener)
			if !ok {
				return fmt.Errorf("%w: %s", registration.ErrNilListener, descriptor.ClassName())
			}
			if err := app.AddContextListener(listener); err != nil {
				return err
			}
		case registration.ListenerRequest:
			listener, ok := descriptor.Listener().(servlet.RequestListener)
			if !ok {
				return fmt.Errorf("%w: %s", registration.ErrNilListener, descriptor.ClassName())
			}
			if err := app.AddRequestListener(listener); err != nil {
				return err
			}
		case registration.ListenerContextAttribute:
			listener, ok := descriptor.Listener().(servlet.ContextAttributeListener)
			if !ok {
				return fmt.Errorf("%w: %s", registration.ErrNilListener, descriptor.ClassName())
			}
			if err := app.AddContextAttributeListener(listener); err != nil {
				return err
			}
		case registration.ListenerRequestAttribute:
			listener, ok := descriptor.Listener().(servlet.RequestAttributeListener)
			if !ok {
				return fmt.Errorf("%w: %s", registration.ErrNilListener, descriptor.ClassName())
			}
			if err := app.AddRequestAttributeListener(listener); err != nil {
				return err
			}
		case registration.ListenerSession:
			continue
		case registration.ListenerSessionAttribute:
			continue
		default:
			return fmt.Errorf("%w: %s", registration.ErrNilListener, descriptor.ClassName())
		}
	}
	return nil
}

func attachRegistrationFilters(deployment *Deployment, start int, nameIndex map[string][]int, descriptors []registration.FilterDescriptor) error {
	orders := make([]orderedFilterBindings, len(deployment.mappings))
	for index := 0; index < start; index++ {
		orders[index].base = cloneFilterBindings(deployment.mappings[index].filterBindings)
	}
	for index := start; index < len(deployment.mappings); index++ {
		orders[index].base = cloneFilterBindings(deployment.mappings[index].filterBindings)
	}
	for _, descriptor := range descriptors {
		if err := attachURLPatternFilters(orders, descriptor); err != nil {
			return err
		}
		if err := attachServletNameFilters(orders, nameIndex, descriptor); err != nil {
			return err
		}
	}
	for index := range deployment.mappings {
		deployment.mappings[index].filterBindings = orders[index].merged()
	}
	return nil
}

func attachURLPatternFilters(orders []orderedFilterBindings, descriptor registration.FilterDescriptor) error {
	for _, mapping := range descriptor.URLPatternMappings() {
		for _, pattern := range mapping.URLPatterns() {
			binding, err := newRegistrationFilterBinding(descriptor, mapping.DispatcherTypes(), pattern)
			if err != nil {
				return err
			}
			for index := range orders {
				orders[index].add(binding, mapping.MatchAfter())
			}
		}
	}
	return nil
}

func attachServletNameFilters(orders []orderedFilterBindings, nameIndex map[string][]int, descriptor registration.FilterDescriptor) error {
	for _, mapping := range descriptor.ServletNameMappings() {
		binding, err := newRegistrationFilterBinding(descriptor, mapping.DispatcherTypes(), "")
		if err != nil {
			return err
		}
		for _, name := range mapping.ServletNames() {
			indexes := nameIndex[name]
			if len(indexes) == 0 {
				return fmt.Errorf("%w: %s", ErrUnknownServletName, name)
			}
			for _, index := range indexes {
				orders[index].add(binding, mapping.MatchAfter())
			}
		}
	}
	return nil
}

func newRegistrationFilterBinding(descriptor registration.FilterDescriptor, dispatchers registration.DispatcherTypes, pattern string) (servlet.FilterBinding, error) {
	return servlet.NewFilterBinding(
		descriptor.Name(),
		descriptor.Filter(),
		servlet.WithFilterDispatchTypes(dispatchers),
		servlet.WithFilterInitParams(descriptor.InitParams()),
		servlet.WithFilterURLPattern(pattern),
	)
}

type orderedFilterBindings struct {
	before []servlet.FilterBinding
	base   []servlet.FilterBinding
	after  []servlet.FilterBinding
}

func (b *orderedFilterBindings) add(binding servlet.FilterBinding, matchAfter bool) {
	if matchAfter {
		b.after = append(b.after, binding)
		return
	}
	b.before = append(b.before, binding)
}

func (b orderedFilterBindings) merged() []servlet.FilterBinding {
	result := make([]servlet.FilterBinding, 0, len(b.before)+len(b.base)+len(b.after))
	result = append(result, b.before...)
	result = append(result, b.base...)
	result = append(result, b.after...)
	return result
}
