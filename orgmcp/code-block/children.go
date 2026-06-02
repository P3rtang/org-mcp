package codeblock

import (
	"fmt"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

var ERR = fmt.Errorf(NO_CHILDREN, "CodeBlock")

func (c *CodeBlock) AddChildren(...Render) error    { return ERR }
func (c *CodeBlock) RemoveChildren(...Uid) error    { return ERR }
func (c *CodeBlock) Children() []Render             { return []Render{} }
func (c *CodeBlock) ChildrenRec(depth int) []Render { return []Render{} }
func (c *CodeBlock) ChildIndentLevel() int          { return c.IndentLevel() + 2 }
func (c *CodeBlock) Insert(int, Render) error       { return ERR }
func (c *CodeBlock) Move(MoveOperation) error       { return ERR }

func (c *CodeBlock) SetParent(parent Render) error {
	c.parent = parent

	c.index = len(parent.Children())

	return nil
}
