package mapper_test

import (
	"strings"
	"testing"

	"structlens/internal/inference"
	"structlens/internal/mapper"
	"structlens/internal/model"
	"structlens/internal/parser"
)

func TestMappingPipelineFromJSONToTables(t *testing.T) {
	input := `{
		"order": {
			"id": 1,
			"customer": {
				"name": "John"
			},
			"items": [
				{ "product": "A", "qty": 2 },
				{ "product": "B", "qty": 1 }
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

	orderSchema := schema.Children["order"]
	if orderSchema == nil {
		t.Fatalf("missing order schema")
	}

	tables, err := mapper.MapSchema(orderSchema, mapper.MapperConfig{FlattenThreshold: 1})
	if err != nil {
		t.Fatalf("map failed: %v", err)
	}

	orders := findTable(tables, "orders")
	if orders == nil {
		t.Fatalf("missing orders table")
	}
	items := findTable(tables, "items")
	if items == nil {
		t.Fatalf("missing items table")
	}
	if !hasColumn(*items, "order_id") {
		t.Fatalf("items table missing order_id")
	}
	if !hasColumn(*items, "product") || !hasColumn(*items, "qty") {
		t.Fatalf("items table missing expected columns")
	}
	for _, table := range tables {
		if table.PrimaryKey != "id" {
			t.Fatalf("table %s primary key = %q, want id", table.Name, table.PrimaryKey)
		}
		if !hasColumn(table, "id") {
			t.Fatalf("table %s missing id column", table.Name)
		}
	}
	if !hasColumn(*orders, "customer_name") {
		t.Fatalf("orders table missing flattened customer_name column")
	}
}
