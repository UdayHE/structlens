package inference_test

import (
	"testing"

	"structlens/internal/inference"
	"structlens/internal/model"
)

func TestMergeNodes_MergesSamplesAndTypes(t *testing.T) {
	sample1 := &model.Node{
		Name: "root",
		Type: model.NodeTypeObject,
		Path: "$",
		Children: []*model.Node{
			{Name: "id", Type: model.NodeTypeNumber, Path: "$.id"},
			{Name: "name", Type: model.NodeTypeString, Path: "$.name"},
		},
	}

	sample2 := &model.Node{
		Name: "root",
		Type: model.NodeTypeObject,
		Path: "$",
		Children: []*model.Node{
			{Name: "id", Type: model.NodeTypeString, Path: "$.id"},
		},
	}

	schema, err := inference.MergeNodes([]*model.Node{sample1, sample2}, inference.InferenceConfig{})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	if schema.OccurrenceCount != 2 {
		t.Fatalf("root occurrence = %d, want 2", schema.OccurrenceCount)
	}
	idSchema := schema.Children["id"]
	if idSchema == nil {
		t.Fatalf("missing id schema")
	}
	if len(idSchema.Types) != 2 {
		t.Fatalf("id types size = %d, want 2", len(idSchema.Types))
	}
	if idSchema.OccurrenceCount != 2 {
		t.Fatalf("id occurrence = %d, want 2", idSchema.OccurrenceCount)
	}
	nameSchema := schema.Children["name"]
	if nameSchema == nil || nameSchema.OccurrenceCount != 1 {
		t.Fatalf("name occurrence invalid: %#v", nameSchema)
	}
}

func TestMergeNodes_ArrayChildrenDeduplicated(t *testing.T) {
	sample := &model.Node{
		Name: "root",
		Type: model.NodeTypeArray,
		Path: "$",
		Children: []*model.Node{
			{
				Name: "0",
				Type: model.NodeTypeObject,
				Path: "$[0]",
				Children: []*model.Node{
					{Name: "id", Type: model.NodeTypeNumber, Path: "$[0].id"},
				},
			},
			{
				Name: "1",
				Type: model.NodeTypeObject,
				Path: "$[1]",
				Children: []*model.Node{
					{Name: "id", Type: model.NodeTypeNumber, Path: "$[1].id"},
				},
			},
		},
		IsArray: true,
	}

	schema, err := inference.MergeNodes([]*model.Node{sample}, inference.InferenceConfig{})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	if len(schema.Children) != 1 {
		t.Fatalf("root children size = %d, want 1", len(schema.Children))
	}
	item := schema.Children["item"]
	if item == nil {
		t.Fatalf("missing item schema")
	}
	if item.OccurrenceCount != 2 {
		t.Fatalf("item occurrence = %d, want 2", item.OccurrenceCount)
	}
	if item.Path != "root.item[]" {
		t.Fatalf("item path = %s, want root.item[]", item.Path)
	}
}
