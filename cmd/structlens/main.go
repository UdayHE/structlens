package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"structlens/internal/export"
	"structlens/internal/inference"
	"structlens/internal/mapper"
	"structlens/internal/model"
	"structlens/internal/parser"
	"structlens/internal/view"
)

const version = "v0.1.0"

type cliConfig struct {
	FlattenThreshold int
	ArrayItemName    string
	View             string
}

type cliRequest struct {
	Config      cliConfig
	InputPath   string
	ShowHelp    bool
	ShowVersion bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	request, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if request.ShowHelp {
		fmt.Fprint(stdout, usageText())
		return 0
	}
	if request.ShowVersion {
		fmt.Fprintln(stdout, version)
		return 0
	}

	output, err := generateOutputFromFile(request.InputPath, request.Config)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintln(stdout, output)
	return 0
}

func parseArgs(args []string) (cliRequest, error) {
	fs := flag.NewFlagSet("structlens", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	request := cliRequest{}
	fs.IntVar(&request.Config.FlattenThreshold, "flatten-threshold", 2, "flatten nested objects with up to this many fields")
	fs.StringVar(&request.Config.ArrayItemName, "array-item-name", "item", "logical name for inferred array items")
	fs.StringVar(&request.Config.View, "view", "sql", "output view: sql or tree")
	fs.BoolVar(&request.ShowVersion, "version", false, "print StructLens version")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			request.ShowHelp = true
			return request, nil
		}
		return cliRequest{}, err
	}

	if request.ShowVersion {
		return request, nil
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return cliRequest{}, errors.New("missing input file.\n\n" + usageText())
	}
	if len(remaining) > 1 {
		return cliRequest{}, errors.New("expected exactly one input file.\n\n" + usageText())
	}

	request.InputPath = remaining[0]
	if request.Config.View != "sql" && request.Config.View != "tree" {
		return cliRequest{}, fmt.Errorf("unsupported view %q. Supported views: sql, tree", request.Config.View)
	}
	return request, nil
}

func generateOutputFromFile(inputPath string, config cliConfig) (string, error) {
	schema, err := generateSchemaFromFile(inputPath, config)
	if err != nil {
		return "", err
	}

	if config.View == "tree" {
		return view.PrintTreeWithArrayItemName(selectMappingRoot(schema), config.ArrayItemName), nil
	}

	tables, err := mapper.MapSchema(selectMappingRoot(schema), mapper.MapperConfig{
		FlattenThreshold: config.FlattenThreshold,
	})
	if err != nil {
		return "", fmt.Errorf("map relational schema for %q: %w", inputPath, err)
	}

	sql, err := export.GenerateSQL(tables)
	if err != nil {
		return "", fmt.Errorf("generate SQL for %q: %w", inputPath, err)
	}

	return sql, nil
}

func generateSQLFromFile(inputPath string, config cliConfig) (string, error) {
	config.View = "sql"
	return generateOutputFromFile(inputPath, config)
}

func generateSchemaFromFile(inputPath string, config cliConfig) (*model.SchemaNode, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("open input file %q: %w", inputPath, err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	if _, err := reader.Peek(1); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("input file %q is empty. Provide a JSON or XML document", inputPath)
		}
		return nil, fmt.Errorf("read input file %q: %w", inputPath, err)
	}

	inputParser, err := parserForPath(inputPath)
	if err != nil {
		return nil, err
	}

	parsedRoot, err := inputParser.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("parse input file %q: %w. Check that the file contains valid %s", inputPath, err, formatLabel(inputPath))
	}

	schema, err := inference.MergeNodes([]*model.Node{parsedRoot}, inference.InferenceConfig{
		ArrayItemName: config.ArrayItemName,
	})
	if err != nil {
		return nil, fmt.Errorf("infer schema for %q: %w", inputPath, err)
	}

	if err := inference.MarkOptionalFields(schema, 1); err != nil {
		return nil, fmt.Errorf("mark optional fields for %q: %w", inputPath, err)
	}

	return schema, nil
}

func parserForPath(inputPath string) (parser.Parser, error) {
	switch strings.ToLower(filepath.Ext(inputPath)) {
	case ".json":
		return parser.NewJSONParser(), nil
	case ".xml":
		return parser.NewXMLParser(), nil
	default:
		return nil, fmt.Errorf("unsupported file format for %q. Supported formats: .json, .xml", inputPath)
	}
}

func selectMappingRoot(schema *model.SchemaNode) *model.SchemaNode {
	if schema == nil {
		return nil
	}
	if schema.Name != "root" || len(schema.Children) != 1 {
		return schema
	}
	for _, child := range schema.Children {
		return child
	}
	return schema
}

func formatLabel(inputPath string) string {
	switch strings.ToLower(filepath.Ext(inputPath)) {
	case ".json":
		return "JSON"
	case ".xml":
		return "XML"
	default:
		return "input"
	}
}

func usageText() string {
	return `StructLens: Convert structured data to SQL schema instantly

Usage:
  structlens [flags] <input.json|input.xml>

Flags:
  --flatten-threshold int
      flatten nested objects with up to this many fields (default 2)
  --array-item-name string
      logical name for inferred array items (default "item")
  --view string
      output view: sql or tree (default "sql")
  --version
      print StructLens version
  --help
      show this help output

Examples:
  structlens examples/simple.json
  structlens --flatten-threshold 1 examples/nested.json
  structlens --view tree examples/nested.json
  structlens --array-item-name entry examples/complex.xml
`
}
