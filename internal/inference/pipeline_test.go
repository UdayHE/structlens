package inference_test

import (
	"strings"
	"testing"

	"structlens/internal/inference"
	"structlens/internal/model"
	"structlens/internal/parser"
)

func TestFullInferencePipeline(t *testing.T) {
	input := `{
		"order": {
			"id": 1,
			"items": [
				{"product": "A", "qty": 2},
				{"product": "B", "qty": 1}
			]
		}
	}`

	rootNode, err := parser.NewJSONParser().Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	schema, err := inference.MergeNodes([]*model.Node{rootNode}, inference.InferenceConfig{})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if err := inference.MarkOptionalFields(schema, 1); err != nil {
		t.Fatalf("mark optional failed: %v", err)
	}

	order := schema.Children["order"]
	if order == nil {
		t.Fatalf("missing order field")
	}
	items := order.Children["items"]
	if items == nil || !items.IsArray {
		t.Fatalf("items not detected as array: %#v", items)
	}
	if items.Path != "root.order.items[]" {
		t.Fatalf("items path = %s, want root.order.items[]", items.Path)
	}

	item := items.Children["item"]
	if item == nil {
		t.Fatalf("missing item schema under items")
	}
	product := item.Children["product"]
	if product == nil {
		t.Fatalf("missing product schema")
	}
	qty := item.Children["qty"]
	if qty == nil {
		t.Fatalf("missing qty schema")
	}
	if qty.ResolvedType() != model.NodeTypeNumber {
		t.Fatalf("qty resolved type = %s, want %s", qty.ResolvedType(), model.NodeTypeNumber)
	}
}
