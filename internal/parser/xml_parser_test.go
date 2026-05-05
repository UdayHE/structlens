package parser_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"structlens/internal/model"
	"structlens/internal/parser"
)

func TestXMLParser_ParseNestedWithAttributes(t *testing.T) {
	input := `<book id="bk-1001"><title>StructLens</title><meta><author active="true">Uday</author></meta></book>`

	parsed, err := parser.NewXMLParser().Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if parsed.Name != "book" {
		t.Fatalf("root name = %s, want book", parsed.Name)
	}
	if parsed.Path != "/book" {
		t.Fatalf("root path = %s, want /book", parsed.Path)
	}
	if parsed.Attributes["id"] != "bk-1001" {
		t.Fatalf("root attribute id = %q, want bk-1001", parsed.Attributes["id"])
	}

	title := childByName(parsed, "title")
	if title == nil || title.Type != model.NodeTypeString {
		t.Fatalf("title node invalid: %#v", title)
	}
	if title.Path != "/book/title" {
		t.Fatalf("title path = %s, want /book/title", title.Path)
	}

	meta := childByName(parsed, "meta")
	if meta == nil || meta.Type != model.NodeTypeObject {
		t.Fatalf("meta node invalid: %#v", meta)
	}
	author := childByName(meta, "author")
	if author == nil || author.Attributes["active"] != "true" {
		t.Fatalf("author node invalid: %#v", author)
	}
}

func TestXMLParser_ParseLargeDeepStream(t *testing.T) {
	const depth = 1200
	reader := deepXMLStream(depth)

	parsed, err := parser.NewXMLParser().Parse(reader)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if parsed.Name != "n0" {
		t.Fatalf("root name = %s, want n0", parsed.Name)
	}

	current := parsed
	for i := 1; i < depth; i++ {
		if len(current.Children) != 1 {
			t.Fatalf("node %s expected one child, got %d", current.Name, len(current.Children))
		}
		current = current.Children[0]
		expectedPath := expectedDeepPath(i)
		if current.Path != expectedPath {
			t.Fatalf("path at depth %d = %s, want %s", i, current.Path, expectedPath)
		}
	}
}

func deepXMLStream(depth int) io.Reader {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()

		for i := 0; i < depth; i++ {
			if _, err := pw.Write([]byte(fmt.Sprintf("<n%d>", i))); err != nil {
				return
			}
		}

		if _, err := pw.Write([]byte("value")); err != nil {
			return
		}

		for i := depth - 1; i >= 0; i-- {
			if _, err := pw.Write([]byte(fmt.Sprintf("</n%d>", i))); err != nil {
				return
			}
		}
	}()
	return pr
}

func expectedDeepPath(level int) string {
	var b strings.Builder
	for i := 0; i <= level; i++ {
		b.WriteString("/n")
		b.WriteString(fmt.Sprintf("%d", i))
	}
	return b.String()
}
