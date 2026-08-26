package container

import (
	"fmt"

	"goark.dev/arkarta/servlet"
)

type filterInitialization struct {
	name      string
	filter    servlet.ManagedFilter
	initParam map[string]string
}

func (d *Deployment) filterInitializations() []filterInitialization {
	result := make([]filterInitialization, 0)
	for _, mapping := range d.mappings {
		for index, binding := range mapping.FilterBindings() {
			target, ok := binding.Filter().(servlet.ManagedFilter)
			if !ok {
				continue
			}
			name := binding.Name()
			if name == "" {
				name = fmt.Sprintf("%s#filter%d", mapping.Name(), index)
			}
			result = append(result, filterInitialization{
				name:      name,
				filter:    target,
				initParam: binding.InitParams(),
			})
		}
	}
	return result
}
