package parser_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"structlens/internal/model"
	"structlens/internal/parser"
)

func TestJSONParser_ParseNested(t *testing.T) {
	input := `{"id":1,"user":{"name":"Ada","active":true},"tags":["a","b"],"meta":null}`

	parsed, err := parser.NewJSONParser().Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if parsed.Type != model.NodeTypeObject {
		t.Fatalf("root type = %s, want %s", parsed.Type, model.NodeTypeObject)
	}
	if parsed.Path != "$" {
		t.Fatalf("root path = %s, want $", parsed.Path)
	}

	idNode := childByName(parsed, "id")
	if idNode == nil || idNode.Type != model.NodeTypeNumber {
		t.Fatalf("id node invalid: %#v", idNode)
	}

	userNode := childByName(parsed, "user")
	if userNode == nil || userNode.Type != model.NodeTypeObject {
		t.Fatalf("user node invalid: %#v", userNode)
	}
	nameNode := childByName(userNode, "name")
	if nameNode == nil || nameNode.Type != model.NodeTypeString {
		t.Fatalf("user.name invalid: %#v", nameNode)
	}

	tagsNode := childByName(parsed, "tags")
	if tagsNode == nil || tagsNode.Type != model.NodeTypeArray || !tagsNode.IsArray {
		t.Fatalf("tags node invalid: %#v", tagsNode)
	}
	if len(tagsNode.Children) != 2 {
		t.Fatalf("tags children = %d, want 2", len(tagsNode.Children))
	}
	if tagsNode.Children[0].Path != "$.tags[0]" {
		t.Fatalf("tags[0] path = %s, want $.tags[0]", tagsNode.Children[0].Path)
	}
}

func TestJSONParser_ParseTopLevelPrimitive(t *testing.T) {
	parsed, err := parser.NewJSONParser().Parse(strings.NewReader(`true`))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parsed.Type != model.NodeTypeBoolean {
		t.Fatalf("type = %s, want %s", parsed.Type, model.NodeTypeBoolean)
	}
	if parsed.Path != "$" {
		t.Fatalf("path = %s, want $", parsed.Path)
	}
}

func TestJSONParser_ParseLargeStream(t *testing.T) {
	reader := largeJSONArrayStream(20000)

	parsed, err := parser.NewJSONParser().Parse(reader)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if parsed.Type != model.NodeTypeArray {
		t.Fatalf("type = %s, want %s", parsed.Type, model.NodeTypeArray)
	}
	if len(parsed.Children) != 20000 {
		t.Fatalf("children = %d, want 20000", len(parsed.Children))
	}
}

func childByName(node *model.Node, name string) *model.Node {
	for _, child := range node.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

func largeJSONArrayStream(size int) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if _, err := pw.Write([]byte("[")); err != nil {
			return
		}

		for i := 0; i < size; i++ {
			if i > 0 {
				if _, err := pw.Write([]byte(",")); err != nil {
					return
				}
			}
			if _, err := pw.Write([]byte(fmt.Sprintf(`{"id":%d,"name":"row"}`, i))); err != nil {
				return
			}
		}

		_, _ = pw.Write([]byte("]"))
	}()
	return pr
}
