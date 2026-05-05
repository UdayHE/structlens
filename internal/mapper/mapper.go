package mapper

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"structlens/internal/model"
)

type MapperConfig struct {
	FlattenThreshold int
}

type mappingContext struct {
	tablesByPath map[string]*model.Table
	order        []*model.Table
}

func MapSchema(root *model.SchemaNode, config MapperConfig) ([]model.Table, error) {
	if root == nil {
		return nil, fmt.Errorf("root schema is nil")
	}
	config = withDefaults(config)
	ctx := newMappingContext()
	rootTable := ensureTable(ctx, root.Path, tableNameForNode(root))
	rootTable.EnsurePrimaryKey()
	if err := mapObjectChildren(root, rootTable, ctx, config); err != nil {
		return nil, err
	}
	return materializeTables(ctx), nil
}

func withDefaults(config MapperConfig) MapperConfig {
	if config.FlattenThreshold < 0 {
		config.FlattenThreshold = 0
	}
	return config
}

func newMappingContext() *mappingContext {
	return &mappingContext{tablesByPath: make(map[string]*model.Table)}
}

func ensureTable(ctx *mappingContext, path, tableName string) *model.Table {
	if table, ok := ctx.tablesByPath[path]; ok {
		return table
	}
	table := &model.Table{Name: tableName}
	table.EnsurePrimaryKey()
	ctx.tablesByPath[path] = table
	ctx.order = append(ctx.order, table)
	return table
}

func materializeTables(ctx *mappingContext) []model.Table {
	result := make([]model.Table, 0, len(ctx.order))
	for _, table := range ctx.order {
		result = append(result, *table)
	}
	return result
}

func mapObjectChildren(node *model.SchemaNode, table *model.Table, ctx *mappingContext, config MapperConfig) error {
	for _, child := range orderedChildren(node) {
		if isPrimitiveNode(child) {
			addPrimitiveColumn(table, child, "")
			continue
		}
		if child.IsArray {
			if err := mapArrayNode(child, table, ctx, config); err != nil {
				return err
			}
			continue
		}
		if shouldFlattenObject(child, config) {
			flattenObjectIntoTable(child, table, config, snakeCase(child.Name))
			continue
		}
		if err := mapNestedObjectAsTable(child, table, ctx, config); err != nil {
			return err
		}
	}
	return nil
}

func mapArrayNode(arrayNode *model.SchemaNode, parentTable *model.Table, ctx *mappingContext, config MapperConfig) error {
	arrayTable := ensureTable(ctx, arrayNode.Path, tableNameForNode(arrayNode))
	addParentReference(arrayTable, parentTable.Name)
	for _, child := range arrayNode.Children {
		if isPrimitiveNode(child) {
			addPrimitiveColumn(arrayTable, child, "")
			continue
		}
		if err := mapObjectChildren(child, arrayTable, ctx, config); err != nil {
			return err
		}
	}
	return nil
}

func mapNestedObjectAsTable(child *model.SchemaNode, parentTable *model.Table, ctx *mappingContext, config MapperConfig) error {
	childTable := ensureTable(ctx, child.Path, tableNameForNode(child))
	addParentReference(childTable, parentTable.Name)
	return mapObjectChildren(child, childTable, ctx, config)
}

func shouldFlatten(node *model.SchemaNode, config MapperConfig) bool {
	return len(node.Children) <= config.FlattenThreshold
}

func shouldFlattenObject(node *model.SchemaNode, config MapperConfig) bool {
	isComplex := len(node.Children) > config.FlattenThreshold
	isRepeated := node.OccurrenceCount > 1
	return !(isComplex && isRepeated)
}

func flattenObjectIntoTable(node *model.SchemaNode, table *model.Table, config MapperConfig, prefix string) {
	for _, child := range orderedChildren(node) {
		columnPrefix := prefix + "_"
		if isPrimitiveNode(child) {
			addPrimitiveColumn(table, child, columnPrefix)
			continue
		}
		if !child.IsArray && shouldFlatten(child, config) {
			flattenObjectIntoTable(child, table, config, columnPrefix+snakeCase(child.Name))
		}
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
	slices.SortFunc(children, func(left, right *model.SchemaNode) int {
		if left == nil && right == nil {
			return 0
		}
		if left == nil {
			return -1
		}
		if right == nil {
			return 1
		}
		if left.Name != right.Name {
			return strings.Compare(left.Name, right.Name)
		}
		return strings.Compare(left.Path, right.Path)
	})
	return children
}

func addPrimitiveColumn(table *model.Table, node *model.SchemaNode, prefix string) {
	name := prefix + snakeCase(node.Name)
	table.AddColumn(name, sqlTypeForNode(node.ResolvedType()), node.Optional)
}

func addParentReference(table *model.Table, parentTable string) {
	parentKey := singularName(parentTable) + "_id"
	table.AddColumn(parentKey, "BIGINT", false)
	table.AddForeignKey(parentKey, parentTable, "id")
}

func isPrimitiveNode(node *model.SchemaNode) bool {
	if node == nil {
		return false
	}
	return len(node.Children) == 0
}

func tableNameForNode(node *model.SchemaNode) string {
	return pluralize(snakeCase(node.Name))
}

func pluralize(value string) string {
	if strings.HasSuffix(value, "s") {
		return value
	}
	return value + "s"
}

func singularName(value string) string {
	if strings.HasSuffix(value, "s") && len(value) > 1 {
		return value[:len(value)-1]
	}
	return value
}

func sqlTypeForNode(nodeType model.NodeType) string {
	switch nodeType {
	case model.NodeTypeBoolean:
		return "BOOLEAN"
	case model.NodeTypeNumber:
		return "DOUBLE PRECISION"
	case model.NodeTypeNull:
		return "TEXT"
	default:
		return "TEXT"
	}
}

func snakeCase(value string) string {
	var b strings.Builder
	for i, r := range value {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteRune('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		if r == '-' || r == ' ' {
			b.WriteRune('_')
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
