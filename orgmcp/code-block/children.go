package codeblock

import (
	"fmt"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

func (c *CodeBlock) Children() []Render             { return []Render{} }
func (c *CodeBlock) ChildrenRec(depth int) []Render { return []Render{} }
func (c *CodeBlock) ChildIndentLevel() int          { return c.IndentLevel() + 2 }
func (c *CodeBlock) AddChildren(...Render) error    { return fmt.Errorf(NO_CHILDREN, c, c.Uid()) }
func (c *CodeBlock) RemoveChildren(...Uid) error    { return fmt.Errorf(NO_CHILDREN, c, c.Uid()) }
func (c *CodeBlock) Insert(int, Render) error       { return fmt.Errorf(NO_CHILDREN, c, c.Uid()) }
func (c *CodeBlock) Move(MoveOperation) error       { return fmt.Errorf(NO_CHILDREN, c, c.Uid()) }

func (c *CodeBlock) SetParent(parent Render) error {
	c.parent = parent

	c.index = len(parent.Children())

	return nil
}
