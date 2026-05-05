package export_test

import (
	"strings"
	"testing"

	"structlens/internal/export"
	"structlens/internal/inference"
	"structlens/internal/mapper"
	"structlens/internal/model"
	"structlens/internal/parser"
)

func TestGenerateSQLOrdersTablesByDependencies(t *testing.T) {
	items := model.Table{
		Name:       "items",
		PrimaryKey: "id",
		Columns: []model.Column{
			{Name: "id", Type: "BIGINT"},
			{Name: "order_id", Type: "BIGINT"},
			{Name: "product", Type: "TEXT"},
		},
		ForeignKeys: []model.ForeignKey{
			{Column: "order_id", RefTable: "orders", RefColumn: "id"},
		},
	}
	orders := model.Table{
		Name:       "orders",
		PrimaryKey: "id",
		Columns: []model.Column{
			{Name: "id", Type: "BIGINT"},
			{Name: "name", Type: "TEXT"},
		},
	}

	sql, err := export.GenerateSQL([]model.Table{items, orders})
	if err != nil {
		t.Fatalf("generate sql failed: %v", err)
	}

	ordersIndex := strings.Index(sql, "CREATE TABLE orders")
	itemsIndex := strings.Index(sql, "CREATE TABLE items")
	if ordersIndex == -1 || itemsIndex == -1 {
		t.Fatalf("missing expected tables in SQL:\n%s", sql)
	}
	if ordersIndex > itemsIndex {
		t.Fatalf("orders table should be created before items:\n%s", sql)
	}
}

func TestGenerateSQLFullPipeline(t *testing.T) {
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

	sql, err := export.GenerateSQL(tables)
	if err != nil {
		t.Fatalf("generate sql failed: %v", err)
	}

	expectedSnippets := []string{
		"CREATE TABLE orders",
		"CREATE TABLE items",
		"PRIMARY KEY",
		"FOREIGN KEY (order_id) REFERENCES orders(id)",
	}
	for _, snippet := range expectedSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("sql missing %q:\n%s", snippet, sql)
		}
	}
}
