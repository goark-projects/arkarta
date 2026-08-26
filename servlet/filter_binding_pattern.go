package servlet

func matchFilterURLPattern(path, pattern string) bool {
	if pattern == "" {
		return true
	}
	kind, value, err := parseMappingPattern(pattern)
	if err != nil {
		return false
	}
	if path == "" {
		path = "/"
	}
	switch kind {
	case mappingDefault:
		return true
	case mappingExact:
		return path == value
	case mappingPrefix:
		return matchPrefix(path, value)
	case mappingExtension:
		return extensionOf(path) == value
	default:
		return false
	}
}
