package view_test

import (
	"strings"
	"testing"

	"structlens/internal/model"
	"structlens/internal/view"
)

func TestPrintTree(t *testing.T) {
	root := model.NewSchemaNode("order", "order")
	root.AddType(model.NodeTypeObject)

	customer := root.AddChild("customer", "order.customer")
	customer.AddType(model.NodeTypeObject)

	name := customer.AddChild("name", "order.customer.name")
	name.AddType(model.NodeTypeString)

	id := root.AddChild("id", "order.id")
	id.AddType(model.NodeTypeNumber)
	id.Optional = true

	items := root.AddChild("items", "order.items[]")
	items.IsArray = true
	items.AddType(model.NodeTypeArray)

	item := items.AddChild("item", "order.items[].item")
	item.AddType(model.NodeTypeObject)

	product := item.AddChild("product", "order.items[].item.product")
	product.AddType(model.NodeTypeString)

	qty := item.AddChild("qty", "order.items[].item.qty")
	qty.AddType(model.NodeTypeNumber)

	got := view.PrintTree(root)
	want := "Schema Summary:\n- Total fields: 7\n- Arrays: 1\n- Optional fields: 1\n\nRoot: order (3 fields)\n├── id (int, optional)\n├── customer (1 fields)\n│   └── name (string)\n└── items[] (array, 2 fields)\n    ├── product (string)\n    └── qty (int)"

	if got != want {
		t.Fatalf("tree output mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestPrintTreeWithArrayItemName(t *testing.T) {
	root := model.NewSchemaNode("feed", "feed")
	root.AddType(model.NodeTypeObject)

	entries := root.AddChild("entries", "feed.entries[]")
	entries.IsArray = true
	entries.AddType(model.NodeTypeArray)

	entry := entries.AddChild("entry", "feed.entries[].entry")
	entry.AddType(model.NodeTypeObject)

	name := entry.AddChild("name", "feed.entries[].entry.name")
	name.AddType(model.NodeTypeString)

	got := view.PrintTreeWithArrayItemName(root, "entry")
	if strings.Contains(got, "└── entry") {
		t.Fatalf("array item node should be flattened visually:\n%s", got)
	}
	if !strings.Contains(got, "Root: feed (1 fields)") || !strings.Contains(got, "entries[] (array, 1 fields)") || !strings.Contains(got, "name (string)") {
		t.Fatalf("tree output missing expected content:\n%s", got)
	}
}
