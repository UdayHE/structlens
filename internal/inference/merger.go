package inference

import (
	"fmt"
	"strconv"
	"strings"

	"structlens/internal/model"
)

// MergeNodes converts multiple parsed Node trees into one unified schema tree.
func MergeNodes(nodes []*model.Node, config InferenceConfig) (*model.SchemaNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes provided")
	}
	return mergeAllNodes(nodes, withDefaults(config))
}

func mergeAllNodes(nodes []*model.Node, config InferenceConfig) (*model.SchemaNode, error) {
	root := model.NewSchemaNode(nodes[0].Name, nodes[0].Name)
	for _, node := range nodes {
		if err := mergeRootNode(node, root, config); err != nil {
			return nil, err
		}
	}
	return root, nil
}

func mergeRootNode(node *model.Node, root *model.SchemaNode, config InferenceConfig) error {
	if node == nil {
		return nil
	}
	return mergeNodeIntoSchema(node, root, config)
}

func mergeNodeIntoSchema(node *model.Node, schema *model.SchemaNode, config InferenceConfig) error {
	if node == nil || schema == nil {
		return nil
	}

	applyNodeMetadata(node, schema)
	for _, child := range node.Children {
		if err := mergeChildNode(node, schema, child, config); err != nil {
			return err
		}
	}
	mergeAttributesIntoSchema(node, schema)
	return nil
}

func mergeAttributesIntoSchema(node *model.Node, schema *model.SchemaNode) {
	for attrName, attrValue := range node.Attributes {
		childName := "@" + attrName
		childPath := BuildPath(schema.Path, childName, false)
		child := schema.AddChild(childName, childPath)
		child.MarkObserved()
		child.AddType(inferAttrType(attrValue))
	}
}

func inferAttrType(value string) model.NodeType {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "true" || lower == "false" {
		return model.NodeTypeBoolean
	}
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return model.NodeTypeNumber
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return model.NodeTypeNumber
	}
	return model.NodeTypeString
}

func applyNodeMetadata(node *model.Node, schema *model.SchemaNode) {
	schema.MarkObserved()
	schema.IsArray = schema.IsArray || node.IsArray || node.Type == model.NodeTypeArray
	schema.AddType(node.Type)
}

func mergeChildNode(parentNode *model.Node, parentSchema *model.SchemaNode, child *model.Node, config InferenceConfig) error {
	if child == nil {
		return nil
	}
	childName, childPath := normalizeChildIdentity(parentNode, parentSchema, child, config)
	childSchema := parentSchema.AddChild(childName, childPath)
	return mergeNodeIntoSchema(child, childSchema, config)
}

func normalizeChildIdentity(parentNode *model.Node, parentSchema *model.SchemaNode, child *model.Node, config InferenceConfig) (string, string) {
	isArrayItem := false
	name := child.Name
	if parentNode.IsArray || parentNode.Type == model.NodeTypeArray {
		if _, err := strconv.Atoi(child.Name); err == nil {
			name = config.ArrayItemName
			isArrayItem = true
		}
	}
	return name, BuildPath(parentSchema.Path, name, isArrayItem || child.IsArray || child.Type == model.NodeTypeArray)
}
