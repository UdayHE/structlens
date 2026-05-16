package view

import "structlens/internal/model"

type AnalysisResponse struct {
	SchemaTree []AnalysisTreeNode `json:"schemaTree"`
	SQL        string             `json:"sql"`
	Metadata   AnalysisMetadata   `json:"metadata"`
	Records    []RecordGroup      `json:"records"`
}

type AnalysisTreeNode struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Path       string             `json:"path"`
	Type       string             `json:"type"`
	Optional   bool               `json:"optional"`
	IsArray    bool               `json:"isArray"`
	ChildCount int                `json:"childCount"`
	Children   []AnalysisTreeNode `json:"children,omitempty"`
}

type AnalysisMetadata struct {
	TotalFields    int `json:"totalFields"`
	ArrayFields    int `json:"arrayFields"`
	OptionalFields int `json:"optionalFields"`
	TableCount     int `json:"tableCount"`
}

func BuildAnalysisResponse(node *model.SchemaNode, sql string, tableCount int, arrayItemName string) AnalysisResponse {
	if node == nil {
		return AnalysisResponse{
			SchemaTree: nil,
			SQL:        sql,
			Metadata: AnalysisMetadata{
				TableCount: tableCount,
			},
		}
	}

	stats := collectStats(node, arrayItemName)
	return AnalysisResponse{
		SchemaTree: []AnalysisTreeNode{buildAnalysisTreeNode(node, arrayItemName)},
		SQL:        sql,
		Metadata: AnalysisMetadata{
			TotalFields:    stats.totalFields,
			ArrayFields:    stats.arrayFields,
			OptionalFields: stats.optionalFields,
			TableCount:     tableCount,
		},
	}
}

func nodeDisplayType(node *model.SchemaNode, visualKids []*model.SchemaNode) string {
	if node.IsArray {
		return string(model.NodeTypeArray)
	}
	if len(visualKids) > 0 {
		return string(model.NodeTypeObject)
	}
	return string(node.ResolvedType())
}

func buildAnalysisTreeNode(node *model.SchemaNode, arrayItemName string) AnalysisTreeNode {
	children := visualChildren(node, arrayItemName)
	result := AnalysisTreeNode{
		ID:         node.Path,
		Name:       node.Name,
		Path:       node.Path,
		Type:       nodeDisplayType(node, children),
		Optional:   node.Optional,
		IsArray:    node.IsArray,
		ChildCount: len(children),
	}

	if len(children) == 0 {
		return result
	}

	result.Children = make([]AnalysisTreeNode, 0, len(children))
	for _, child := range children {
		result.Children = append(result.Children, buildAnalysisTreeNode(child, arrayItemName))
	}

	return result
}
