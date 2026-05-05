package inference_test

import (
	"testing"

	"structlens/internal/inference"
	"structlens/internal/model"
)

func TestMarkOptionalFields(t *testing.T) {
	root := model.NewSchemaNode("root", "$")
	root.OccurrenceCount = 3

	id := root.AddChild("id", "$.id")
	id.OccurrenceCount = 3

	name := root.AddChild("name", "$.name")
	name.OccurrenceCount = 2

	if err := inference.MarkOptionalFields(root, 3); err != nil {
		t.Fatalf("mark optional failed: %v", err)
	}

	if root.Optional {
		t.Fatalf("root optional = true, want false")
	}
	if id.Optional {
		t.Fatalf("id optional = true, want false")
	}
	if !name.Optional {
		t.Fatalf("name optional = false, want true")
	}
}
