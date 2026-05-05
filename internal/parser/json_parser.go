package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"structlens/internal/model"
)

// JSONParser parses JSON streams into the unified Node model.
type JSONParser struct{}

// NewJSONParser creates a Parser for JSON input.
func NewJSONParser() Parser {
	return JSONParser{}
}

// Parse consumes JSON from reader using a streaming decoder.
func (p JSONParser) Parse(reader io.Reader) (*model.Node, error) {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()

	root, err := parseJSONValue(decoder, "root", "$")
	if err != nil {
		return nil, err
	}

	if token, err := decoder.Token(); err == nil {
		return nil, fmt.Errorf("invalid JSON: unexpected extra token %v", token)
	}
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return root, nil
}

func parseJSONValue(decoder *json.Decoder, name, path string) (*model.Node, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("read token at %s: %w", path, err)
	}

	switch value := token.(type) {
	case json.Delim:
		return parseDelimitedValue(decoder, value, name, path)
	default:
		return buildScalarNode(name, path, value)
	}
}

func parseDelimitedValue(decoder *json.Decoder, delim json.Delim, name, path string) (*model.Node, error) {
	switch delim {
	case '{':
		return parseJSONObject(decoder, name, path)
	case '[':
		return parseJSONArray(decoder, name, path)
	default:
		return nil, fmt.Errorf("invalid JSON at %s: unexpected delimiter %q", path, delim)
	}
}

func parseJSONObject(decoder *json.Decoder, name, path string) (*model.Node, error) {
	node := &model.Node{Name: name, Type: model.NodeTypeObject, Path: path}
	for decoder.More() {
		if err := parseObjectField(decoder, node, path); err != nil {
			return nil, err
		}
	}
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("close object at %s: %w", path, err)
	}
	return node, nil
}

func parseJSONArray(decoder *json.Decoder, name, path string) (*model.Node, error) {
	node := &model.Node{Name: name, Type: model.NodeTypeArray, IsArray: true, Path: path}
	for index := 0; decoder.More(); index++ {
		if err := parseArrayElement(decoder, node, path, index); err != nil {
			return nil, err
		}
	}
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("close array at %s: %w", path, err)
	}
	return node, nil
}

func parseObjectField(decoder *json.Decoder, node *model.Node, path string) error {
	key, err := readObjectKey(decoder, path)
	if err != nil {
		return err
	}
	childPath := objectChildPath(path, key)
	child, err := parseJSONValue(decoder, key, childPath)
	if err != nil {
		return err
	}
	node.Children = append(node.Children, child)
	return nil
}

func readObjectKey(decoder *json.Decoder, path string) (string, error) {
	keyToken, err := decoder.Token()
	if err != nil {
		return "", fmt.Errorf("read object key at %s: %w", path, err)
	}
	key, ok := keyToken.(string)
	if !ok {
		return "", fmt.Errorf("invalid JSON at %s: non-string key %v", path, keyToken)
	}
	return key, nil
}

func parseArrayElement(decoder *json.Decoder, node *model.Node, path string, index int) error {
	childName := strconv.Itoa(index)
	childPath := arrayChildPath(path, index)
	child, err := parseJSONValue(decoder, childName, childPath)
	if err != nil {
		return err
	}
	node.Children = append(node.Children, child)
	return nil
}

func buildScalarNode(name, path string, value any) (*model.Node, error) {
	switch value.(type) {
	case string:
		return &model.Node{Name: name, Type: model.NodeTypeString, Path: path}, nil
	case json.Number, float64:
		return &model.Node{Name: name, Type: model.NodeTypeNumber, Path: path}, nil
	case bool:
		return &model.Node{Name: name, Type: model.NodeTypeBoolean, Path: path}, nil
	case nil:
		return &model.Node{Name: name, Type: model.NodeTypeNull, Path: path}, nil
	default:
		return nil, fmt.Errorf("unsupported JSON value at %s: %T", path, value)
	}
}

func objectChildPath(parentPath, key string) string {
	if parentPath == "$" {
		return "$." + key
	}
	return parentPath + "." + key
}

func arrayChildPath(parentPath string, index int) string {
	return parentPath + "[" + strconv.Itoa(index) + "]"
}
