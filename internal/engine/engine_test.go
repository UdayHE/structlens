package engine

import (
	"strings"
	"testing"
)

func TestAnalyzeReaderBuildsSchemaAndSQL(t *testing.T) {
	result, err := AnalyzeReader("order.json", strings.NewReader(`{
		"order": {
			"id": 1,
			"customer": {"name": "Ada"},
			"items": [{"product": "A", "qty": 2}]
		}
	}`), Config{
		FlattenThreshold: 1,
		ArrayItemName:    "item",
	})
	if err != nil {
		t.Fatalf("analyze reader failed: %v", err)
	}

	if result.MappingRoot == nil || result.MappingRoot.Name != "order" {
		t.Fatalf("unexpected mapping root: %#v", result.MappingRoot)
	}
	if !strings.Contains(result.SQL, `CREATE TABLE "orders"`) {
		t.Fatalf("expected sql output, got:\n%s", result.SQL)
	}
	if result.Metadata.TotalFields == 0 || result.Metadata.TableCount == 0 {
		t.Fatalf("unexpected metadata: %+v", result.Metadata)
	}
}

func TestAnalyzeReaderRejectsUnsupportedFormat(t *testing.T) {
	_, err := AnalyzeReader("order.yaml", strings.NewReader("order: 1"), Config{})
	if err == nil {
		t.Fatal("expected unsupported format error")
	}
	if !strings.Contains(err.Error(), "Supported formats: .json, .xml") {
		t.Fatalf("unexpected error: %v", err)
	}
}
