package container

import "goark.dev/arkarta/servlet"

func cloneProfiles(src []Profile) []Profile {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Profile, len(src))
	copy(dst, src)
	return dst
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func cloneFilters(src []servlet.Filter) []servlet.Filter {
	if len(src) == 0 {
		return nil
	}
	dst := make([]servlet.Filter, len(src))
	copy(dst, src)
	return dst
}

func cloneMappings(src []Mapping) []Mapping {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Mapping, len(src))
	copy(dst, src)
	for i := range dst {
		dst[i].filters = cloneFilters(dst[i].filters)
		dst[i].initParam = cloneStringMap(dst[i].initParam)
	}
	return dst
}
