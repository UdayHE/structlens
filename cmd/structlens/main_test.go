package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSQLFromFileJSON(t *testing.T) {
	inputPath := filepath.Join("..", "..", "examples", "order.json")

	sql, err := generateSQLFromFile(inputPath, cliConfig{
		FlattenThreshold: 1,
		ArrayItemName:    "item",
	})
	if err != nil {
		t.Fatalf("generate sql from file failed: %v", err)
	}

	expectedSnippets := []string{
		`CREATE TABLE "orders"`,
		`CREATE TABLE "items"`,
		"PRIMARY KEY",
		`FOREIGN KEY ("order_id") REFERENCES "orders"("id")`,
	}
	for _, snippet := range expectedSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("sql missing %q:\n%s", snippet, sql)
		}
	}
}

func TestRunWritesSQLToStdout(t *testing.T) {
	inputPath := filepath.Join("..", "..", "examples", "order.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--flatten-threshold", "1", inputPath}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), `CREATE TABLE "orders"`) {
		t.Fatalf("stdout missing orders table:\n%s", stdout.String())
	}
}

func TestRunWritesTreeToStdout(t *testing.T) {
	inputPath := filepath.Join("..", "..", "examples", "order.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--view", "tree", inputPath}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Schema Summary:") || !strings.Contains(stdout.String(), "Root: order") || !strings.Contains(stdout.String(), "items[] (array, 2 fields)") {
		t.Fatalf("stdout missing tree output:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "\n    └── item") {
		t.Fatalf("tree output should flatten redundant item node:\n%s", stdout.String())
	}
}

func TestRunWritesJSONToStdout(t *testing.T) {
	inputPath := filepath.Join("..", "..", "examples", "order.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--view", "json", inputPath}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	var payload struct {
		SchemaTree []struct {
			Name string `json:"name"`
		} `json:"schemaTree"`
		SQL      string `json:"sql"`
		Metadata struct {
			TotalFields int `json:"totalFields"`
			TableCount  int `json:"tableCount"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("json output should decode: %v\n%s", err, stdout.String())
	}
	if len(payload.SchemaTree) == 0 || payload.SchemaTree[0].Name != "order" {
		t.Fatalf("unexpected schema tree payload: %+v", payload.SchemaTree)
	}
	if !strings.Contains(payload.SQL, `CREATE TABLE "orders"`) {
		t.Fatalf("json output missing SQL payload:\n%s", payload.SQL)
	}
	if payload.Metadata.TotalFields == 0 || payload.Metadata.TableCount == 0 {
		t.Fatalf("unexpected metadata payload: %+v", payload.Metadata)
	}
}

func TestRunHelpOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--help"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Usage:") || !strings.Contains(stdout.String(), "Examples:") {
		t.Fatalf("help output missing expected sections:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "--view string") {
		t.Fatalf("help output missing view flag:\n%s", stdout.String())
	}
}

func TestRunVersionOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != version {
		t.Fatalf("version output = %q, want %q", strings.TrimSpace(stdout.String()), version)
	}
}

func TestRunMissingInputShowsHelpfulMessage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(nil, &stdout, &stderr)

	if exitCode == 0 {
		t.Fatal("expected non-zero exit code for missing input")
	}
	if !strings.Contains(stderr.String(), "missing input file") || !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("stderr missing expected guidance:\n%s", stderr.String())
	}
}

func TestGenerateSQLFromFileUnsupportedFormat(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "order.yaml")
	if err := os.WriteFile(inputPath, []byte("order:\n  id: 1\n"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := generateSQLFromFile(inputPath, cliConfig{})
	if err == nil {
		t.Fatal("expected unsupported format error")
	}
	if !strings.Contains(err.Error(), "Supported formats: .json, .xml") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseArgsRejectsUnsupportedView(t *testing.T) {
	_, err := parseArgs([]string{"--view", "graph", "input.json"})
	if err == nil {
		t.Fatal("expected unsupported view error")
	}
	if !strings.Contains(err.Error(), "Supported views: sql, tree, json") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateSQLFromFileEmptyInput(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "empty.json")
	if err := os.WriteFile(inputPath, nil, 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := generateSQLFromFile(inputPath, cliConfig{})
	if err == nil {
		t.Fatal("expected empty file error")
	}
	if !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateSQLFromFileInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "broken.json")
	if err := os.WriteFile(inputPath, []byte(`{"order":`), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := generateSQLFromFile(inputPath, cliConfig{})
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
	if !strings.Contains(err.Error(), "valid JSON") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateSQLFromFileDeterministicOutput(t *testing.T) {
	inputPath := filepath.Join("..", "..", "examples", "order.json")

	first, err := generateSQLFromFile(inputPath, cliConfig{
		FlattenThreshold: 1,
		ArrayItemName:    "item",
	})
	if err != nil {
		t.Fatalf("first generate failed: %v", err)
	}
	second, err := generateSQLFromFile(inputPath, cliConfig{
		FlattenThreshold: 1,
		ArrayItemName:    "item",
	})
	if err != nil {
		t.Fatalf("second generate failed: %v", err)
	}
	if first != second {
		t.Fatalf("SQL output is not deterministic\nfirst:\n%s\n\nsecond:\n%s", first, second)
	}
}

func TestGenerateSQLFromFileInvalidXML(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "broken.xml")
	if err := os.WriteFile(inputPath, []byte(`<catalog><book></catalog>`), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := generateSQLFromFile(inputPath, cliConfig{})
	if err == nil {
		t.Fatal("expected invalid XML error")
	}
	if !strings.Contains(err.Error(), "valid XML") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateSQLFromFileDeepXML(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "deep.xml")
	if err := os.WriteFile(inputPath, []byte(buildDeepXML(250)), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	sql, err := generateSQLFromFile(inputPath, cliConfig{})
	if err != nil {
		t.Fatalf("generate sql from deep XML failed: %v", err)
	}
	if !strings.Contains(sql, `CREATE TABLE "n0s"`) {
		t.Fatalf("unexpected SQL output for deep XML:\n%s", sql)
	}
}

func buildDeepXML(depth int) string {
	var b strings.Builder
	for i := 0; i < depth; i++ {
		b.WriteString(fmt.Sprintf("<n%d>", i))
	}
	b.WriteString("value")
	for i := depth - 1; i >= 0; i-- {
		b.WriteString(fmt.Sprintf("</n%d>", i))
	}
	return b.String()
}
