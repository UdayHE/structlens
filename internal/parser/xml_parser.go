package parser

import (
	"encoding/xml"
	"fmt"
	"io"

	"structlens/internal/model"
)

// XMLParser parses XML streams into the unified Node model.
type XMLParser struct{}

type xmlParseState struct {
	root       *model.Node
	stack      []*model.Node
	textBuffer []string
}

// NewXMLParser creates a Parser for XML input.
func NewXMLParser() Parser {
	return XMLParser{}
}

// Parse consumes XML from reader using a streaming decoder.
func (p XMLParser) Parse(reader io.Reader) (*model.Node, error) {
	decoder := xml.NewDecoder(reader)
	state := newXMLParseState()
	for {
		token, done, err := readNextXMLToken(decoder)
		if err != nil {
			return nil, err
		}
		if done {
			break
		}
		if err := handleXMLToken(token, state); err != nil {
			return nil, err
		}
	}
	return validateXMLParseState(state)
}

func newXMLParseState() *xmlParseState {
	return &xmlParseState{
		stack:      make([]*model.Node, 0, 32),
		textBuffer: make([]string, 0, 4),
	}
}

func readNextXMLToken(decoder *xml.Decoder) (xml.Token, bool, error) {
	token, err := decoder.Token()
	if err == io.EOF {
		return nil, true, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("parse XML token: %w", err)
	}
	return token, false, nil
}

func handleXMLToken(token xml.Token, state *xmlParseState) error {
	switch tok := token.(type) {
	case xml.StartElement:
		handleStartElement(tok, state)
	case xml.CharData:
		handleCharData(tok, state)
	case xml.EndElement:
		return handleEndElement(tok, state)
	}
	return nil
}

func handleStartElement(start xml.StartElement, state *xmlParseState) {
	node := newXMLNode(start, state.stack)
	appendToParent(node, state)
	state.stack = append(state.stack, node)
	state.textBuffer = append(state.textBuffer, "")
}

func newXMLNode(start xml.StartElement, stack []*model.Node) *model.Node {
	return &model.Node{
		Name:       start.Name.Local,
		Type:       model.NodeTypeObject,
		Path:       xmlPath(stack, start.Name.Local),
		Attributes: xmlAttributes(start.Attr),
	}
}

func appendToParent(node *model.Node, state *xmlParseState) {
	if len(state.stack) == 0 {
		state.root = node
		return
	}
	parent := state.stack[len(state.stack)-1]
	parent.Children = append(parent.Children, node)
}

func handleCharData(charData xml.CharData, state *xmlParseState) {
	if len(state.stack) == 0 || len(state.textBuffer) == 0 {
		return
	}
	state.textBuffer[len(state.textBuffer)-1] += string(charData)
}

func handleEndElement(end xml.EndElement, state *xmlParseState) error {
	if len(state.stack) == 0 || len(state.textBuffer) == 0 {
		return fmt.Errorf("invalid XML: unexpected end element %s", end.Name.Local)
	}
	current := state.stack[len(state.stack)-1]
	text := state.textBuffer[len(state.textBuffer)-1]
	assignLeafType(current, text)
	popXMLState(state)
	return nil
}

func assignLeafType(node *model.Node, text string) {
	if len(node.Children) > 0 {
		return
	}
	if isOnlyWhitespaceOrEmpty(text) {
		node.Type = model.NodeTypeNull
		return
	}
	node.Type = model.NodeTypeString
}

func popXMLState(state *xmlParseState) {
	state.stack = state.stack[:len(state.stack)-1]
	state.textBuffer = state.textBuffer[:len(state.textBuffer)-1]
}

func validateXMLParseState(state *xmlParseState) (*model.Node, error) {
	if state.root == nil {
		return nil, fmt.Errorf("invalid XML: no root element found")
	}
	if len(state.stack) != 0 {
		return nil, fmt.Errorf("invalid XML: unclosed elements remain")
	}
	return state.root, nil
}

func xmlAttributes(attrs []xml.Attr) map[string]string {
	if len(attrs) == 0 {
		return nil
	}
	result := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		result[attr.Name.Local] = attr.Value
	}
	return result
}

func xmlPath(stack []*model.Node, current string) string {
	if len(stack) == 0 {
		return "/" + current
	}
	return stack[len(stack)-1].Path + "/" + current
}

func isOnlyWhitespaceOrEmpty(value string) bool {
	for _, r := range value {
		switch r {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return false
		}
	}
	return true
}
