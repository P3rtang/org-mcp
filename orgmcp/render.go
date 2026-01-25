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
	SetParent(Render) error
	RemoveChildren(...Uid) error
	Children() []Render
	ChildrenRec() []Render
	Uid() Uid
	ParentUid() Uid
}
