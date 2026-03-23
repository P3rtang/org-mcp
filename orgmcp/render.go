package orgmcp

import (
	"context"
	"strings"

	"github.com/p3rtang/org-mcp/utils/option"
)

type RenderStatus string

func (r RenderStatus) String() string {
	return string(r)
}

type RenderBase interface {
	Movable

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
	ParentUid() Uid
	Status() RenderStatus
	TagList() TagList
	Preview(length int) string
	Path() string
	Insert(context.Context, int, Render) error
	InsertAfter(context.Context, Uid, Render) error
	InsertBefore(context.Context, Uid, Render) error
	Move(context.Context, MoveOperation[Render]) error
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
