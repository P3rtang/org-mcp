package orgmcp

import (
	"github.com/p3rtang/org-mcp/utils/option"
	"strings"
)

type RenderStatus string

func (r RenderStatus) String() string {
	return string(r)
}

type Render interface {
	CheckProgress() option.Option[Progress]
	Render(builder *strings.Builder, depth int)
	IndentLevel() int
	ChildIndentLevel() int
	Level() int
	Location(table map[Uid]int) int
	AddChildren(...Render) error
	SetParent(Render) error
	RemoveChildren(...Uid) error
	Children() []Render
	ChildrenRec(depth int) []Render
	Uid() Uid
	ParentUid() Uid
	Status() RenderStatus
	TagList() TagList
	Preview(length int) string
	Path() string
}
