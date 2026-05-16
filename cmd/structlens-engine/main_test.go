package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunEmitsJSONResponse(t *testing.T) {
	input := `{
		"fileName": "order.json",
		"content": "{\"order\":{\"id\":1,\"items\":[{\"product\":\"A\"}]}}",
		"flattenThreshold": 1,
		"arrayItemName": "item"
	}`

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(strings.NewReader(input), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	var payload struct {
		SQL      string `json:"sql"`
		Metadata struct {
			TableCount int `json:"tableCount"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v\n%s", err, stdout.String())
	}
	if !strings.Contains(payload.SQL, `CREATE TABLE "orders"`) {
		t.Fatalf("missing sql output:\n%s", payload.SQL)
	}
	if payload.Metadata.TableCount == 0 {
		t.Fatalf("expected non-zero table count: %+v", payload.Metadata)
	}
}
