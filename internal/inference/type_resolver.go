package inference

import (
	"strings"

	"structlens/internal/model"
)

type detectedTypes struct {
	seenInt    bool
	seenFloat  bool
	seenString bool
	seenBool   bool
	seenOther  bool
}

// ResolveType resolves mixed detected types into one final type.
func ResolveType(typeSet map[model.NodeType]struct{}) model.NodeType {
	normalized := normalizeTypes(typeSet)
	if len(normalized) == 0 {
		return model.NodeTypeNull
	}
	detected := detectTypes(normalized)
	return resolveDetectedType(detected)
}

func normalizeTypes(typeSet map[model.NodeType]struct{}) []string {
	normalized := make([]string, 0, len(typeSet))
	for nodeType := range typeSet {
		normalized = append(normalized, strings.ToLower(strings.TrimSpace(string(nodeType))))
	}
	return normalized
}

func detectTypes(types []string) detectedTypes {
	result := detectedTypes{}
	for _, t := range types {
		recordType(t, &result)
	}
	return result
}

func recordType(value string, detected *detectedTypes) {
	switch value {
	case "int", "integer":
		detected.seenInt = true
	case "float", "double", "number":
		detected.seenFloat = true
	case "string":
		detected.seenString = true
	case "bool", "boolean":
		detected.seenBool = true
	case "", "null":
		return
	default:
		detected.seenOther = true
	}
}

func resolveDetectedType(d detectedTypes) model.NodeType {
	if resolved, ok := resolveMixedTypes(d); ok {
		return resolved
	}
	if resolved, ok := resolveNumericTypes(d); ok {
		return resolved
	}
	return firstPresentType(d)
}

func resolveMixedTypes(d detectedTypes) (model.NodeType, bool) {
	if d.seenOther || hasStringConflict(d) || hasBoolNumericConflict(d) {
		return model.NodeTypeString, true
	}
	return "", false
}

func resolveNumericTypes(d detectedTypes) (model.NodeType, bool) {
	if d.seenInt && d.seenFloat {
		return model.NodeTypeNumber, true
	}
	return "", false
}

func hasStringConflict(d detectedTypes) bool {
	return d.seenString && (d.seenInt || d.seenFloat || d.seenBool)
}

func hasBoolNumericConflict(d detectedTypes) bool {
	return d.seenBool && (d.seenInt || d.seenFloat)
}

func firstPresentType(d detectedTypes) model.NodeType {
	if d.seenString {
		return model.NodeTypeString
	}
	if d.seenFloat {
		return model.NodeTypeNumber
	}
	if d.seenInt {
		return model.NodeTypeNumber
	}
	if d.seenBool {
		return model.NodeTypeBoolean
	}
	return model.NodeTypeNull
}
