package orgmcp

import (
	"github.com/p3rtang/org-mcp/utils/option"
	"strings"
)

type Render interface {
	CheckProgress() option.Option[Progress]
	Render(builder *strings.Builder, depth int)
	IndentLevel() int
	AddChildren(...Render) error
	RemoveChildren(...int) error
	Children() []Render
	ChildrenRec() []Render
	Uid() int
	ParentUid() int
}
