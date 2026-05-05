package model

import "testing"

func TestTableEnsurePrimaryKeyAddsDefaultID(t *testing.T) {
	table := &Table{Name: "orders"}

	table.EnsurePrimaryKey()

	if table.PrimaryKey != "id" {
		t.Fatalf("primary key = %q, want %q", table.PrimaryKey, "id")
	}
	if len(table.Columns) != 1 {
		t.Fatalf("column count = %d, want 1", len(table.Columns))
	}
	if table.Columns[0].Name != "id" || table.Columns[0].Type != "BIGINT" || table.Columns[0].Nullable {
		t.Fatalf("unexpected primary key column: %#v", table.Columns[0])
	}
}

func TestTableAddColumnSkipsDuplicates(t *testing.T) {
	table := &Table{Name: "orders"}

	table.AddColumn("id", "BIGINT", false)
	table.AddColumn("id", "TEXT", true)

	if len(table.Columns) != 1 {
		t.Fatalf("column count = %d, want 1", len(table.Columns))
	}
	if table.Columns[0].Type != "BIGINT" {
		t.Fatalf("existing column should be preserved, got %#v", table.Columns[0])
	}
}
