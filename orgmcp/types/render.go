package orgmcp_types

import (
	"errors"
	"strings"

	"github.com/p3rtang/org-mcp/utils/option"
)

var (
	HasNoChildren = errors.New("This item cannot have children.")
)

type RenderStatus string

func (r RenderStatus) String() string {
	return string(r)
}

type RenderBase interface {
	Uid() Uid
	ParentUid() Uid
	Level() int
	IndentLevel() int
	Location(table map[Uid]int) int
	Path() string

	Status() RenderStatus
	CheckProgress() option.Option[Progress]
	TagList() TagList

	AddChildren(...Render) error
	SetParent(Render) error
	RemoveChildren(...Uid) error
	Children() []Render
	ChildrenRec(depth int) []Render
	ChildIndentLevel() int
	Insert(int, Render) error
	Move(MoveOperation) error
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

	Preview(length int) string
}
