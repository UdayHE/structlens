package engine

import (
	"bufio"
	"errors"
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

type Config struct {
	FlattenThreshold int
	ArrayItemName    string
}

type Result struct {
	SchemaRoot  *model.SchemaNode
	MappingRoot *model.SchemaNode
	ParsedRoot  *model.Node
	SQL         string
	Metadata    view.AnalysisMetadata
	Tables      []model.Table
}

func AnalyzeFile(inputPath string, config Config) (*Result, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("open input file %q: %w", inputPath, err)
	}
	defer file.Close()

	return AnalyzeReader(inputPath, file, config)
}

func AnalyzeReader(inputName string, reader io.Reader, config Config) (*Result, error) {
	buffered := bufio.NewReader(reader)
	if _, err := buffered.Peek(1); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("input file %q is empty. Provide a JSON or XML document", inputName)
		}
		return nil, fmt.Errorf("read input file %q: %w", inputName, err)
	}

	inputParser, err := parserForPath(inputName)
	if err != nil {
		return nil, err
	}

	parsedRoot, err := inputParser.Parse(buffered)
	if err != nil {
		return nil, fmt.Errorf("parse input file %q: %w. Check that the file contains valid %s", inputName, err, formatLabel(inputName))
	}

	arrayItemName := config.ArrayItemName
	if arrayItemName == "" {
		arrayItemName = "item"
	}

	schema, err := inference.MergeNodes([]*model.Node{parsedRoot}, inference.InferenceConfig{
		ArrayItemName: arrayItemName,
	})
	if err != nil {
		return nil, fmt.Errorf("infer schema for %q: %w", inputName, err)
	}

	if err := inference.MarkOptionalFields(schema, 1); err != nil {
		return nil, fmt.Errorf("mark optional fields for %q: %w", inputName, err)
	}

	mappingRoot := SelectMappingRoot(schema)
	tables, err := mapper.MapSchema(mappingRoot, mapper.MapperConfig{
		FlattenThreshold: config.FlattenThreshold,
	})
	if err != nil {
		return nil, fmt.Errorf("map relational schema for %q: %w", inputName, err)
	}

	sql, err := export.GenerateSQL(tables)
	if err != nil {
		return nil, fmt.Errorf("generate SQL for %q: %w", inputName, err)
	}

	response := view.BuildAnalysisResponse(mappingRoot, sql, len(tables), arrayItemName)
	return &Result{
		SchemaRoot:  schema,
		MappingRoot: mappingRoot,
		ParsedRoot:  parsedRoot,
		SQL:         sql,
		Metadata:    response.Metadata,
		Tables:      tables,
	}, nil
}

func BuildResponse(result *Result, arrayItemName string) view.AnalysisResponse {
	if result == nil {
		return view.AnalysisResponse{}
	}
	resp := view.BuildAnalysisResponse(result.MappingRoot, result.SQL, len(result.Tables), arrayItemName)
	resp.Records = view.BuildRecords(result.ParsedRoot)
	return resp
}

func SelectMappingRoot(schema *model.SchemaNode) *model.SchemaNode {
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
