package orgmcp

import (
	"github.com/p3rtang/org-mcp/utils/option"
	"strings"
)

type RenderStatus string

func (r RenderStatus) String() string {
	return string(r)
}

type RenderBase interface {
	CheckProgress() option.Option[Progress]
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

type RenderOrg interface {
	RenderBase
	Render(builder *strings.Builder, depth int)
}

type RenderMarkdown interface {
	RenderBase
	RenderMarkdown(builder *strings.Builder, depth int)
}

type Render interface {
	RenderMarkdown
	RenderOrg
}
