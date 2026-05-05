package inference_test

import (
	"testing"

	"structlens/internal/inference"
	"structlens/internal/model"
)

func TestResolveType(t *testing.T) {
	tests := []struct {
		name  string
		input map[model.NodeType]struct{}
		want  model.NodeType
	}{
		{name: "number only", input: typeSet(model.NodeTypeNumber), want: model.NodeTypeNumber},
		{name: "string only", input: typeSet(model.NodeTypeString), want: model.NodeTypeString},
		{name: "number and string", input: typeSet(model.NodeTypeNumber, model.NodeTypeString), want: model.NodeTypeString},
		{name: "bool and number", input: typeSet(model.NodeTypeBoolean, model.NodeTypeNumber), want: model.NodeTypeString},
		{name: "bool only", input: typeSet(model.NodeTypeBoolean), want: model.NodeTypeBoolean},
		{name: "null and number", input: typeSet(model.NodeTypeNull, model.NodeTypeNumber), want: model.NodeTypeNumber},
		{name: "empty", input: nil, want: model.NodeTypeNull},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := inference.ResolveType(tc.input)
			if got != tc.want {
				t.Fatalf("ResolveType(%v) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func typeSet(types ...model.NodeType) map[model.NodeType]struct{} {
	result := make(map[model.NodeType]struct{}, len(types))
	for _, nodeType := range types {
		result[nodeType] = struct{}{}
	}
	return result
}
