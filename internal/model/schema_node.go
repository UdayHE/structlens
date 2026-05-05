package model

import "fmt"

var resolveTypeFn = defaultResolveType

func defaultResolveType(typeSet map[NodeType]struct{}) NodeType {
	for nodeType := range typeSet {
		return nodeType
	}
	return NodeTypeNull
}

// SchemaNode represents an inferred schema node aggregated across samples.
type SchemaNode struct {
	Name            string
	Path            string
	Types           map[NodeType]struct{}
	IsArray         bool
	OccurrenceCount int
	Optional        bool
	Children        map[string]*SchemaNode
}

// NewSchemaNode creates an initialized schema node.
func NewSchemaNode(name, path string) *SchemaNode {
	return &SchemaNode{
		Name:     name,
		Path:     path,
		Types:    make(map[NodeType]struct{}, 1),
		Children: make(map[string]*SchemaNode),
	}
}

// AddType records a detected type for this schema node.
func (s *SchemaNode) AddType(nodeType NodeType) {
	if s.Types == nil {
		s.Types = make(map[NodeType]struct{}, 1)
	}
	s.Types[nodeType] = struct{}{}
}

// AddChild returns an existing child by name or creates a new one.
func (s *SchemaNode) AddChild(name, path string) *SchemaNode {
	if s.Children == nil {
		s.Children = make(map[string]*SchemaNode)
	}
	if child, exists := s.Children[name]; exists {
		return child
	}
	child := NewSchemaNode(name, path)
	s.Children[name] = child
	return child
}

// MarkObserved increments occurrence count for this node.
func (s *SchemaNode) MarkObserved() {
	s.OccurrenceCount++
}

// Merge combines another schema node into this node.
func (s *SchemaNode) Merge(other *SchemaNode) error {
	if other == nil {
		return nil
	}
	if s.Name != "" && other.Name != "" && s.Name != other.Name {
		return fmt.Errorf("cannot merge schema nodes with different names: %q vs %q", s.Name, other.Name)
	}
	if s.Path != "" && other.Path != "" && s.Path != other.Path {
		return fmt.Errorf("cannot merge schema nodes with different paths: %q vs %q", s.Path, other.Path)
	}

	if s.Name == "" {
		s.Name = other.Name
	}
	if s.Path == "" {
		s.Path = other.Path
	}

	s.IsArray = s.IsArray || other.IsArray
	s.Optional = s.Optional || other.Optional
	s.OccurrenceCount += other.OccurrenceCount

	if s.Types == nil {
		s.Types = make(map[NodeType]struct{}, len(other.Types))
	}
	for nodeType := range other.Types {
		s.Types[nodeType] = struct{}{}
	}

	if s.Children == nil {
		s.Children = make(map[string]*SchemaNode, len(other.Children))
	}
	for name, otherChild := range other.Children {
		currentChild, exists := s.Children[name]
		if !exists {
			s.Children[name] = otherChild.Clone()
			continue
		}
		if err := currentChild.Merge(otherChild); err != nil {
			return err
		}
	}

	return nil
}

// Clone creates a deep copy of the schema node.
func (s *SchemaNode) Clone() *SchemaNode {
	if s == nil {
		return nil
	}

	copyNode := &SchemaNode{
		Name:            s.Name,
		Path:            s.Path,
		IsArray:         s.IsArray,
		OccurrenceCount: s.OccurrenceCount,
		Optional:        s.Optional,
		Types:           make(map[NodeType]struct{}, len(s.Types)),
		Children:        make(map[string]*SchemaNode, len(s.Children)),
	}

	for nodeType := range s.Types {
		copyNode.Types[nodeType] = struct{}{}
	}
	for name, child := range s.Children {
		copyNode.Children[name] = child.Clone()
	}

	return copyNode
}

// SetTypeResolver registers schema type-resolution behavior.
func SetTypeResolver(fn func(map[NodeType]struct{}) NodeType) {
	if fn == nil {
		resolveTypeFn = defaultResolveType
		return
	}
	resolveTypeFn = fn
}

// ResolvedType returns the final resolved type for this schema node.
func (s *SchemaNode) ResolvedType() NodeType {
	return resolveTypeFn(s.Types)
}
