package codeblock

import (
	"fmt"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/option"
)

func (c *CodeBlock) Uid() Uid {
	return NewUid(c.parent.Uid().String() + c.name.UnwrapOr(fmt.Sprintf("c%d", c.index)))
}

func (c *CodeBlock) ParentUid() Uid       { return c.parent.Uid() }
func (c *CodeBlock) Level() int           { return c.parent.Level() + 1 }
func (c *CodeBlock) IndentLevel() int     { return c.parent.ChildIndentLevel() }
func (c *CodeBlock) Path() string         { return c.parent.Path() + "/" + c.Uid().String() }
func (c *CodeBlock) Status() RenderStatus { return "" }
func (c *CodeBlock) TagList() TagList     { return c.parent.TagList() }

func (c *CodeBlock) CheckProgress() option.Option[Progress] { return option.None[Progress]() }

func (c *CodeBlock) Location(table map[Uid]int) (loc int) {
	if val, ok := table[c.Uid()]; ok {
		return val
	}

	if c.parent == nil {
		return 0
	}

	loc += c.parent.Location(table)

	for i, child := range c.parent.ChildrenRec(-1) {
		if child.Uid() == c.Uid() {
			loc += i + 1
			break
		}
	}

	return
}
