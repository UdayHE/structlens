package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"structlens/internal/engine"
)

type analyzeRequest struct {
	FileName         string `json:"fileName"`
	FilePath         string `json:"filePath"`
	Content          string `json:"content"`
	FlattenThreshold int    `json:"flattenThreshold"`
	ArrayItemName    string `json:"arrayItemName"`
}

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr))
}

func run(stdin io.Reader, stdout, stderr io.Writer) int {
	request, err := decodeRequest(stdin)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	cfg := engine.Config{
		FlattenThreshold: request.FlattenThreshold,
		ArrayItemName:    request.ArrayItemName,
	}
	var result *engine.Result
	if request.FilePath != "" {
		result, err = engine.AnalyzeFile(request.FilePath, cfg)
	} else {
		result, err = engine.AnalyzeReader(request.FileName, strings.NewReader(request.Content), cfg)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	response := engine.BuildResponse(result, request.ArrayItemName)
	if err := json.NewEncoder(stdout).Encode(response); err != nil {
		fmt.Fprintln(stderr, fmt.Errorf("encode engine response: %w", err))
		return 1
	}

	return 0
}

func decodeRequest(stdin io.Reader) (analyzeRequest, error) {
	var request analyzeRequest
	if err := json.NewDecoder(stdin).Decode(&request); err != nil {
		return analyzeRequest{}, fmt.Errorf("decode engine request: %w", err)
	}
	if request.FileName == "" {
		return analyzeRequest{}, fmt.Errorf("missing fileName in engine request")
	}
	if request.FlattenThreshold == 0 {
		request.FlattenThreshold = 2
	}
	if request.ArrayItemName == "" {
		request.ArrayItemName = "item"
	}
	return request, nil
}
