package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"structlens/internal/engine"
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
	fs.StringVar(&request.Config.View, "view", "sql", "output view: sql, tree, json, or records")
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
	switch request.Config.View {
	case "sql", "tree", "json", "records":
	default:
		return cliRequest{}, fmt.Errorf("unsupported view %q. Supported views: sql, tree, json, records", request.Config.View)
	}
	return request, nil
}

func generateOutputFromFile(inputPath string, config cliConfig) (string, error) {
	result, err := engine.AnalyzeFile(inputPath, engine.Config{
		FlattenThreshold: config.FlattenThreshold,
		ArrayItemName:    config.ArrayItemName,
	})
	if err != nil {
		return "", err
	}

	if config.View == "tree" {
		return view.PrintTreeWithArrayItemName(result.MappingRoot, config.ArrayItemName), nil
	}

	if config.View == "records" {
		return view.PrintRecords(result.ParsedRoot), nil
	}

	if config.View == "json" {
		response := engine.BuildResponse(result, config.ArrayItemName)
		payload, err := json.Marshal(response)
		if err != nil {
			return "", fmt.Errorf("encode analysis response for %q: %w", inputPath, err)
		}
		return string(payload), nil
	}

	return result.SQL, nil
}

func generateSQLFromFile(inputPath string, config cliConfig) (string, error) {
	config.View = "sql"
	return generateOutputFromFile(inputPath, config)
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
      output view: sql, tree, json, or records (default "sql")
  --version
      print StructLens version
  --help
      show this help output

Examples:
  structlens examples/simple.json
  structlens --flatten-threshold 1 examples/nested.json
  structlens --view tree examples/nested.json
  structlens --view json examples/nested.json
  structlens --array-item-name entry examples/complex.xml
`
}
