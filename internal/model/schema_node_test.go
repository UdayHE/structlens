package model_test

import (
	"testing"

	"structlens/internal/model"
)

func TestSchemaNodeAddType(t *testing.T) {
	node := model.NewSchemaNode("user", "$.user")
	node.AddType(model.NodeTypeObject)
	node.AddType(model.NodeTypeObject)
	node.AddType(model.NodeTypeNull)

	if len(node.Types) != 2 {
		t.Fatalf("types size = %d, want 2", len(node.Types))
	}
}

func TestSchemaNodeMerge(t *testing.T) {
	left := model.NewSchemaNode("user", "$.user")
	left.AddType(model.NodeTypeObject)
	left.OccurrenceCount = 2
	leftChild := left.AddChild("name", "$.user.name")
	leftChild.AddType(model.NodeTypeString)
	leftChild.OccurrenceCount = 2

	right := model.NewSchemaNode("user", "$.user")
	right.AddType(model.NodeTypeNull)
	right.IsArray = true
	right.Optional = true
	right.OccurrenceCount = 1
	rightChild := right.AddChild("name", "$.user.name")
	rightChild.AddType(model.NodeTypeNull)
	rightChild.OccurrenceCount = 1
	ageChild := right.AddChild("age", "$.user.age")
	ageChild.AddType(model.NodeTypeNumber)
	ageChild.OccurrenceCount = 1

	if err := left.Merge(right); err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	if len(left.Types) != 2 {
		t.Fatalf("types size = %d, want 2", len(left.Types))
	}
	if !left.IsArray {
		t.Fatalf("isArray = false, want true")
	}
	if !left.Optional {
		t.Fatalf("optional = false, want true")
	}
	if left.OccurrenceCount != 3 {
		t.Fatalf("occurrenceCount = %d, want 3", left.OccurrenceCount)
	}
	if len(left.Children) != 2 {
		t.Fatalf("children size = %d, want 2", len(left.Children))
	}
	if left.Children["name"].OccurrenceCount != 3 {
		t.Fatalf("name occurrenceCount = %d, want 3", left.Children["name"].OccurrenceCount)
	}
}
