package view

import (
	"fmt"
	"strings"

	"structlens/internal/model"
)

const defaultArrayItemName = "item"

type treeStats struct {
	totalFields    int
	arrayFields    int
	optionalFields int
}

// PrintTree renders a schema node as an ASCII tree.
func PrintTree(node *model.SchemaNode) string {
	return PrintTreeWithArrayItemName(node, defaultArrayItemName)
}

// PrintTreeWithArrayItemName renders a schema node as an ASCII tree using the
// provided inferred array item name for visual flattening.
func PrintTreeWithArrayItemName(node *model.SchemaNode, arrayItemName string) string {
	if node == nil {
		return ""
	}
	if arrayItemName == "" {
		arrayItemName = defaultArrayItemName
	}

	var b strings.Builder
	stats := collectStats(node, arrayItemName)
	b.WriteString(formatSummary(stats))
	b.WriteString("\n\n")
	b.WriteString("Root: ")
	b.WriteString(nodeLabel(node, arrayItemName))
	b.WriteString("\n")

	children := visualChildren(node, arrayItemName)
	for i, child := range children {
		writeTree(&b, child, "", i == len(children)-1, arrayItemName)
	}

	return strings.TrimRight(b.String(), "\n")
}

func writeTree(b *strings.Builder, node *model.SchemaNode, prefix string, isLast bool, arrayItemName string) {
	if node == nil {
		return
	}

	b.WriteString(prefix)
	if isLast {
		b.WriteString("└── ")
	} else {
		b.WriteString("├── ")
	}
	b.WriteString(nodeLabel(node, arrayItemName))
	b.WriteString("\n")

	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	children := visualChildren(node, arrayItemName)
	for i, child := range children {
		writeTree(b, child, childPrefix, i == len(children)-1, arrayItemName)
	}
}

func nodeLabel(node *model.SchemaNode, arrayItemName string) string {
	name := node.Name
	if node.IsArray {
		return fmt.Sprintf("%s[] (array, %d fields)%s", name, visibleChildCount(node, arrayItemName), optionalSuffix(node))
	}
	if len(visualChildren(node, arrayItemName)) != 0 {
		return fmt.Sprintf("%s (%d fields)%s", name, visibleChildCount(node, arrayItemName), optionalSuffix(node))
	}
	typeLabel := readableType(node)
	if typeLabel == "" {
		return name + optionalSuffix(node)
	}
	return fmt.Sprintf("%s (%s%s)", name, typeLabel, optionalTypeSuffix(node))
}

func readableType(node *model.SchemaNode) string {
	if node == nil {
		return ""
	}
	if len(node.Children) != 0 {
		return ""
	}
	switch node.ResolvedType() {
	case model.NodeTypeNumber:
		return "int"
	case model.NodeTypeString:
		return "string"
	case model.NodeTypeBoolean:
		return "bool"
	default:
		return ""
	}
}

func orderedChildren(node *model.SchemaNode) []*model.SchemaNode {
	if node == nil || len(node.Children) == 0 {
		return nil
	}

	children := make([]*model.SchemaNode, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, child)
	}
	sortNodes(children)
	return children
}

func visualChildren(node *model.SchemaNode, arrayItemName string) []*model.SchemaNode {
	children := orderedChildren(node)
	if node == nil || !node.IsArray || len(children) != 1 {
		return children
	}
	child := children[0]
	if child == nil || child.Name != arrayItemName {
		return children
	}
	return orderedChildren(child)
}

func collectStats(node *model.SchemaNode, arrayItemName string) treeStats {
	if node == nil {
		return treeStats{}
	}

	stats := treeStats{totalFields: 1}
	if node.IsArray {
		stats.arrayFields++
	}
	if node.Optional {
		stats.optionalFields++
	}

	children := orderedChildren(node)
	if node.IsArray && len(children) == 1 && children[0] != nil && children[0].Name == arrayItemName {
		return addStats(stats, collectChildrenStats(children[0], arrayItemName))
	}
	return addStats(stats, collectChildrenStats(node, arrayItemName))
}

func collectChildrenStats(node *model.SchemaNode, arrayItemName string) treeStats {
	total := treeStats{}
	for _, child := range orderedChildren(node) {
		total = addStats(total, collectStats(child, arrayItemName))
	}
	return total
}

func addStats(left, right treeStats) treeStats {
	return treeStats{
		totalFields:    left.totalFields + right.totalFields,
		arrayFields:    left.arrayFields + right.arrayFields,
		optionalFields: left.optionalFields + right.optionalFields,
	}
}

func formatSummary(stats treeStats) string {
	return fmt.Sprintf(
		"Schema Summary:\n- Total fields: %d\n- Arrays: %d\n- Optional fields: %d",
		stats.totalFields,
		stats.arrayFields,
		stats.optionalFields,
	)
}

func optionalSuffix(node *model.SchemaNode) string {
	if node != nil && node.Optional {
		return " (optional)"
	}
	return ""
}

func optionalTypeSuffix(node *model.SchemaNode) string {
	if node != nil && node.Optional {
		return ", optional"
	}
	return ""
}

func visibleChildCount(node *model.SchemaNode, arrayItemName string) int {
	return len(visualChildren(node, arrayItemName))
}

func sortNodes(nodes []*model.SchemaNode) {
	for i := 1; i < len(nodes); i++ {
		current := nodes[i]
		j := i - 1
		for ; j >= 0 && compareNodes(nodes[j], current) > 0; j-- {
			nodes[j+1] = nodes[j]
		}
		nodes[j+1] = current
	}
}

func compareNodes(left, right *model.SchemaNode) int {
	leftRank := nodeGroupRank(left)
	rightRank := nodeGroupRank(right)
	if leftRank != rightRank {
		return leftRank - rightRank
	}
	leftName := ""
	if left != nil {
		leftName = left.Name
	}
	rightName := ""
	if right != nil {
		rightName = right.Name
	}
	if leftName < rightName {
		return -1
	}
	if leftName > rightName {
		return 1
	}
	return 0
}

func nodeGroupRank(node *model.SchemaNode) int {
	if node == nil {
		return 3
	}
	if node.IsArray {
		return 2
	}
	if len(node.Children) == 0 {
		return 0
	}
	return 1
}
