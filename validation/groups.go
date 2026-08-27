package validation

import "strings"

const (
	// DefaultGroup 表示未显式声明分组的默认校验组。
	DefaultGroup = "default"

	defaultGroupTagName = "arkarta-groups"
)

type groupSet map[string]struct{}

func newGroupSet(groups ...string) groupSet {
	if len(groups) == 0 {
		return groupSet{DefaultGroup: {}}
	}
	result := make(groupSet, len(groups))
	for _, group := range groups {
		group = normalizeGroup(group)
		if group != "" {
			result[group] = struct{}{}
		}
	}
	if len(result) == 0 {
		result[DefaultGroup] = struct{}{}
	}
	return result
}

func parseGroupTag(tag string) groupSet {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return groupSet{DefaultGroup: {}}
	}
	if tag == "-" {
		return groupSet{}
	}
	return newGroupSet(strings.FieldsFunc(tag, func(r rune) bool {
		return r == ',' || r == '|' || r == ';' || r == ' '
	})...)
}

func (g groupSet) activeIn(active groupSet) bool {
	if len(g) == 0 || len(active) == 0 {
		return false
	}
	for group := range g {
		if _, ok := active[group]; ok {
			return true
		}
	}
	return false
}

func normalizeGroups(groups []string) []string {
	normalized := newGroupSet(groups...)
	result := make([]string, 0, len(normalized))
	for group := range normalized {
		result = append(result, group)
	}
	return result
}

func normalizeGroup(group string) string {
	return strings.ToLower(strings.TrimSpace(group))
}
