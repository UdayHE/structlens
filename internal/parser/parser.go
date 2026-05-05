package parser

import (
	"io"

	"structlens/internal/model"
)

// Parser defines a streaming parser that transforms structured input into a Node tree.
type Parser interface {
	Parse(reader io.Reader) (*model.Node, error)
}
