package table

import (
	"fmt"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

func (t *Table) Children() (c []Render)         { return }
func (t *Table) ChildrenRec(_ int) (c []Render) { return }
func (t *Table) ChildIndentLevel() int          { return t.parent.ChildIndentLevel() }

func (t *Table) AddChildren(...Render) error { return fmt.Errorf(NO_CHILDREN, t, t.Uid()) }
func (t *Table) RemoveChildren(...Uid) error { return fmt.Errorf(NO_CHILDREN, t, t.Uid()) }
func (t *Table) Insert(int, Render) error    { return fmt.Errorf(NO_CHILDREN, t, t.Uid()) }
func (t *Table) Move(MoveOperation) error    { return fmt.Errorf(NO_CHILDREN, t, t.Uid()) }

func (t *Table) SetParent(p Render) error {
	t.parent = p

	return nil
}
