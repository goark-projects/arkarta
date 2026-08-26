package container

import (
	"sort"

	"goark.dev/arkarta/servlet"
)

type servletInitialization struct {
	index            int
	name             string
	target           servlet.Servlet
	initParam        map[string]string
	loadOnStartup    int
	hasLoadOnStartup bool
}

func (d *Deployment) servletInitializations() []servletInitialization {
	result := make([]servletInitialization, 0)
	seen := make(map[string]struct{})
	for index, mapping := range d.mappings {
		target, ok := mapping.Handler().(servlet.Servlet)
		if !ok {
			continue
		}
		name := mapping.Name()
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		loadOrder, hasLoadOrder := mapping.LoadOnStartup()
		result = append(result, servletInitialization{
			index:            index,
			name:             name,
			target:           target,
			initParam:        mapping.InitParams(),
			loadOnStartup:    loadOrder,
			hasLoadOnStartup: hasLoadOrder,
		})
	}
	sort.SliceStable(result, func(i, j int) bool {
		left := result[i]
		right := result[j]
		if left.hasLoadOnStartup && right.hasLoadOnStartup {
			if left.loadOnStartup == right.loadOnStartup {
				return left.index < right.index
			}
			return left.loadOnStartup < right.loadOnStartup
		}
		if left.hasLoadOnStartup != right.hasLoadOnStartup {
			return left.hasLoadOnStartup
		}
		return left.index < right.index
	})
	return result
}

func (d *Deployment) servletMappings() []Mapping {
	result := make([]Mapping, 0, len(d.mappings))
	for _, mapping := range d.mappings {
		if _, ok := mapping.Handler().(servlet.Servlet); ok {
			result = append(result, mapping)
		}
	}
	return result
}
