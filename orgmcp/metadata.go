package orgmcp

import (
	"errors"
	"main/utils/option"
	"strings"
)

type MetadataString string

const (
	CLOSED MetadataString = "CLOSED"
)

// Metadata represents a metadata or continuation line (like CLOSED:, SCHEDULED:, etc.)
type Metadata struct {
	content  string
	location int

	parent Render
}

// Enforce that Metadata implements the Render interface at compile time
var _ Render = (*Metadata)(nil)

// NewMetadata creates a new Metadata instance from a raw line with indentation
func MetadataFromString(str string, parent Render) Metadata {
	// Count leading spaces to determine indentation
	content := strings.TrimSpace(str)

	return Metadata{content: content, parent: parent}
}

func (m *Metadata) Render(builder *strings.Builder) {
	builder.WriteString(strings.Repeat(" ", m.IndentLevel()))

	// Render with the original indentation level
	builder.WriteString(m.content)
}

func (m *Metadata) CheckProgress() option.Option[Progress] {
	return option.None[Progress]()
}

func (m *Metadata) Location() int {
	return m.location
}

func (m *Metadata) IndentLevel() int {
	return m.parent.IndentLevel()
}

func (m *Metadata) AddChild(_ Render) error {
	return errors.New("Metadata does not support children")
}

func (m *Metadata) Children() []Render {
	return []Render{}
}
