package inference

// BuildPath creates deterministic schema paths without numeric indexes.
func BuildPath(parentPath, name string, isArray bool) string {
	segment := name
	if isArray {
		segment += "[]"
	}
	if parentPath == "" {
		return segment
	}
	return parentPath + "." + segment
}
