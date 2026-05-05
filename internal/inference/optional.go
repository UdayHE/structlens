package inference

import (
	"fmt"

	"structlens/internal/model"
)

// MarkOptionalFields marks schema nodes as optional when they appear in fewer than totalSamples.
func MarkOptionalFields(root *model.SchemaNode, totalSamples int) error {
	if root == nil {
		return fmt.Errorf("root schema is nil")
	}
	if totalSamples <= 0 {
		return fmt.Errorf("totalSamples must be > 0")
	}

	markOptionalRecursive(root, totalSamples)
	return nil
}

func markOptionalRecursive(node *model.SchemaNode, totalSamples int) {
	node.Optional = node.OccurrenceCount < totalSamples
	for _, child := range node.Children {
		markOptionalRecursive(child, totalSamples)
	}
}
