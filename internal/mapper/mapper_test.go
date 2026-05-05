package mapper_test

import (
	"testing"

	"structlens/internal/mapper"
	"structlens/internal/model"
)

func TestMapSchema_BasicObjectAndArray(t *testing.T) {
	root := model.NewSchemaNode("order", "order")
	items := root.AddChild("items", "order.items[]")
	items.IsArray = true
	item := items.AddChild("item", "order.items[].item")
	product := item.AddChild("product", "order.items[].item.product")
	product.AddType(model.NodeTypeString)
	qty := item.AddChild("qty", "order.items[].item.qty")
	qty.AddType(model.NodeTypeNumber)

	tables, err := mapper.MapSchema(root, mapper.MapperConfig{FlattenThreshold: 1})
	if err != nil {
		t.Fatalf("map failed: %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("table count = %d, want 2", len(tables))
	}

	order := findTable(tables, "orders")
	if order == nil {
		t.Fatalf("missing orders table")
	}
	if order.PrimaryKey != "id" || !hasColumn(*order, "id") {
		t.Fatalf("orders table missing id primary key: %#v", order)
	}
	itemsTable := findTable(tables, "items")
	if itemsTable == nil {
		t.Fatalf("missing items table")
	}
	if itemsTable.PrimaryKey != "id" || !hasColumn(*itemsTable, "id") {
		t.Fatalf("items table missing id primary key: %#v", itemsTable)
	}
	if !hasColumn(*itemsTable, "order_id") {
		t.Fatalf("items table missing order_id")
	}
	if !hasColumn(*itemsTable, "product") || !hasColumn(*itemsTable, "qty") {
		t.Fatalf("items table missing expected columns")
	}
}

func TestMapSchema_FlattenObject(t *testing.T) {
	root := model.NewSchemaNode("user", "user")
	profile := root.AddChild("profile", "user.profile")
	name := profile.AddChild("name", "user.profile.name")
	name.AddType(model.NodeTypeString)
	age := profile.AddChild("age", "user.profile.age")
	age.AddType(model.NodeTypeNumber)

	tables, err := mapper.MapSchema(root, mapper.MapperConfig{FlattenThreshold: 2})
	if err != nil {
		t.Fatalf("map failed: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("table count = %d, want 1", len(tables))
	}
	user := tables[0]
	if user.PrimaryKey != "id" || !hasColumn(user, "id") {
		t.Fatalf("users table missing id primary key: %#v", user)
	}
	if !hasColumn(user, "profile_name") || !hasColumn(user, "profile_age") {
		t.Fatalf("flattened columns missing in users table")
	}
}

func TestMapSchema_FlattensNonRepeatedComplexObject(t *testing.T) {
	root := model.NewSchemaNode("order", "order")
	root.MarkObserved()
	customer := root.AddChild("customer", "order.customer")
	customer.MarkObserved()
	name := customer.AddChild("name", "order.customer.name")
	name.AddType(model.NodeTypeString)
	email := customer.AddChild("email", "order.customer.email")
	email.AddType(model.NodeTypeString)

	tables, err := mapper.MapSchema(root, mapper.MapperConfig{FlattenThreshold: 1})
	if err != nil {
		t.Fatalf("map failed: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("table count = %d, want 1", len(tables))
	}

	order := tables[0]
	if !hasColumn(order, "customer_name") || !hasColumn(order, "customer_email") {
		t.Fatalf("customer fields were not flattened: %#v", order.Columns)
	}
}

func TestMapSchema_CreatesTableForRepeatedComplexObject(t *testing.T) {
	root := model.NewSchemaNode("order", "order")
	root.MarkObserved()
	address := root.AddChild("address", "order.address")
	address.OccurrenceCount = 2
	street := address.AddChild("street", "order.address.street")
	street.AddType(model.NodeTypeString)
	city := address.AddChild("city", "order.address.city")
	city.AddType(model.NodeTypeString)

	tables, err := mapper.MapSchema(root, mapper.MapperConfig{FlattenThreshold: 1})
	if err != nil {
		t.Fatalf("map failed: %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("table count = %d, want 2", len(tables))
	}

	var childTable *model.Table
	for i := range tables {
		if tables[i].Name != "orders" {
			childTable = &tables[i]
			break
		}
	}
	if childTable == nil {
		t.Fatalf("missing child table for repeated complex object")
	}
	if !hasColumn(*childTable, "order_id") {
		t.Fatalf("child table missing order_id")
	}
}

func findTable(tables []model.Table, name string) *model.Table {
	for i := range tables {
		if tables[i].Name == name {
			return &tables[i]
		}
	}
	return nil
}

func hasColumn(table model.Table, name string) bool {
	for _, col := range table.Columns {
		if col.Name == name {
			return true
		}
	}
	return false
}
