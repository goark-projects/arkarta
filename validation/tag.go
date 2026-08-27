package validation

import (
	"strings"
	"unicode"
)

const defaultTagName = "arkarta"

func parseRules(tag string) ([]Rule, error) {
	tag = strings.TrimSpace(tag)
	if tag == "" || tag == "-" {
		return nil, nil
	}
	parts := strings.Split(tag, ",")
	rules := make([]Rule, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, param, _ := strings.Cut(part, "=")
		name = strings.ToLower(strings.TrimSpace(name))
		param = strings.TrimSpace(param)
		if !validRuleName(name) {
			return nil, ErrInvalidRule
		}
		rules = append(rules, Rule{name: name, param: param})
	}
	return rules, nil
}

func fieldPath(parent, name string) string {
	if parent == "" {
		return name
	}
	if name == "" {
		return parent
	}
	return parent + "." + name
}

func fieldName(fieldName, jsonTag string) string {
	if jsonTag == "-" {
		return ""
	}
	name, _, _ := strings.Cut(jsonTag, ",")
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}
	if fieldName == "" {
		return ""
	}
	runes := []rune(fieldName)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func validRuleName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}
