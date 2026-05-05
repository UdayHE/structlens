package model

// NodeType represents a normalized data type for structured content.
type NodeType string

const (
	NodeTypeObject  NodeType = "object"
	NodeTypeArray   NodeType = "array"
	NodeTypeString  NodeType = "string"
	NodeTypeNumber  NodeType = "number"
	NodeTypeBoolean NodeType = "boolean"
	NodeTypeNull    NodeType = "null"
)

// Node is the unified structure used for both JSON and XML content.
// JSON object fields and XML child elements are represented in Children.
// XML attributes are represented in Attributes.
type Node struct {
	Name       string            `json:"name"`
	Type       NodeType          `json:"type"`
	Children   []*Node           `json:"children,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	IsArray    bool              `json:"is_array,omitempty"`
	Path       string            `json:"path"`
}
