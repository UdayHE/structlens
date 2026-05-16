package view

import (
	"fmt"
	"sort"
	"strings"

	"structlens/internal/model"
)

// RecordGroup contains all instances of a single entity type from the source file.
type RecordGroup struct {
	TypeName  string           `json:"typeName"`
	Instances []RecordInstance `json:"instances"`
}

// RecordInstance is a single occurrence of an entity type with its attribute values.
type RecordInstance struct {
	Attributes map[string]string  `json:"attributes"`
	Children   []RecordChildGroup `json:"children,omitempty"`
}

// RecordChildGroup summarizes child elements nested inside a record instance.
type RecordChildGroup struct {
	Name      string   `json:"name"`
	Count     int      `json:"count"`
	KeyValues []string `json:"keyValues,omitempty"`
}

const maxInstancesPerGroup = 500

// BuildRecords extracts actual data instances from the raw parsed node tree, grouped by entity type.
// Nodes with XML attributes are treated as entity instances; nodes without attributes are transparent
// containers (e.g. <Add>, <Change>) that are walked through.
// Each group is capped at maxInstancesPerGroup instances to keep the response payload bounded.
func BuildRecords(root *model.Node) []RecordGroup {
	if root == nil {
		return nil
	}
	start := rawMappingRoot(root)
	if start == nil {
		return nil
	}

	byName := make(map[string]*RecordGroup)
	var order []string

	var walk func(*model.Node)
	walk = func(n *model.Node) {
		if n == nil {
			return
		}
		if len(n.Attributes) > 0 {
			if _, exists := byName[n.Name]; !exists {
				byName[n.Name] = &RecordGroup{TypeName: n.Name}
				order = append(order, n.Name)
			}
			g := byName[n.Name]
			if len(g.Instances) < maxInstancesPerGroup {
				g.Instances = append(g.Instances, buildRecordInstance(n))
			}
			return
		}
		for _, child := range n.Children {
			walk(child)
		}
	}

	for _, child := range start.Children {
		walk(child)
	}

	groups := make([]RecordGroup, 0, len(order))
	for _, name := range order {
		groups = append(groups, *byName[name])
	}
	return groups
}

// PrintRecords formats record groups as human-readable text for CLI output.
func PrintRecords(root *model.Node) string {
	groups := BuildRecords(root)
	if len(groups) == 0 {
		return "(no records found — records view is supported for XML files with attribute-based entities)\n"
	}

	var b strings.Builder
	for i, g := range groups {
		if i > 0 {
			b.WriteString("\n")
		}
		suffix := "records"
		if len(g.Instances) == 1 {
			suffix = "record"
		}
		header := fmt.Sprintf("%s (%d %s)", g.TypeName, len(g.Instances), suffix)
		fmt.Fprintf(&b, "%s\n%s\n", header, strings.Repeat("─", len([]rune(header))))
		for j, inst := range g.Instances {
			printRecordInstance(&b, inst, j+1)
		}
	}
	return b.String()
}

func rawMappingRoot(root *model.Node) *model.Node {
	if root.Name != "root" || len(root.Children) != 1 {
		return root
	}
	return root.Children[0]
}

func buildRecordInstance(n *model.Node) RecordInstance {
	inst := RecordInstance{Attributes: n.Attributes}

	childMap := make(map[string][]map[string]string)
	var childOrder []string
	for _, child := range n.Children {
		if _, exists := childMap[child.Name]; !exists {
			childOrder = append(childOrder, child.Name)
		}
		childMap[child.Name] = append(childMap[child.Name], child.Attributes)
	}

	if len(childOrder) > 0 {
		inst.Children = make([]RecordChildGroup, 0, len(childOrder))
		for _, childName := range childOrder {
			attrSets := childMap[childName]
			keyVals := make([]string, 0, len(attrSets))
			for _, a := range attrSets {
				if kv := nameValueOf(a); kv != "" {
					keyVals = append(keyVals, kv)
				}
			}
			inst.Children = append(inst.Children, RecordChildGroup{
				Name:      childName,
				Count:     len(attrSets),
				KeyValues: keyVals,
			})
		}
	}

	return inst
}

func nameValueOf(attrs map[string]string) string {
	if v, ok := attrs["name"]; ok && v != "" {
		return v
	}
	var nameKeys []string
	for k := range attrs {
		if strings.HasSuffix(k, "Name") {
			nameKeys = append(nameKeys, k)
		}
	}
	if len(nameKeys) > 0 {
		sort.Strings(nameKeys)
		if v := attrs[nameKeys[0]]; v != "" {
			return v
		}
	}
	return ""
}

func printRecordInstance(b *strings.Builder, inst RecordInstance, num int) {
	keys := make([]string, 0, len(inst.Attributes))
	for k := range inst.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	const perLine = 3
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		v := inst.Attributes[k]
		if v == "" {
			v = `""`
		}
		pairs = append(pairs, k+"="+v)
	}

	indent := "       "
	for i := 0; i < len(pairs); i += perLine {
		end := i + perLine
		if end > len(pairs) {
			end = len(pairs)
		}
		if i == 0 {
			fmt.Fprintf(b, "  [%d]  %s\n", num, strings.Join(pairs[i:end], "  "))
		} else {
			fmt.Fprintf(b, "%s%s\n", indent, strings.Join(pairs[i:end], "  "))
		}
	}

	for _, cg := range inst.Children {
		if len(cg.KeyValues) > 0 {
			fmt.Fprintf(b, "%s↳ %s ×%d: %s\n", indent, cg.Name, cg.Count, strings.Join(cg.KeyValues, ", "))
		} else {
			fmt.Fprintf(b, "%s↳ %s ×%d\n", indent, cg.Name, cg.Count)
		}
	}

	b.WriteString("\n")
}
