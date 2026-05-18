package table

import (
	"fmt"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

func (t *Table) Uid() Uid {
	return NewUid(fmt.Sprintf("%s.%s", t.parent.Uid(), t.uid))
}

func (t *Table) ParentUid() Uid {
	return t.parent.Uid()
}

func (t *Table) Level() int {
	return t.parent.Level() + 1
}

func (t *Table) IndentLevel() int {
	return t.parent.ChildIndentLevel()
}

func (t *Table) Location(table map[Uid]int) (loc int) {
	if val, ok := table[t.Uid()]; ok {
		return val
	}

	loc += t.parent.Location(table)

	for i, child := range t.parent.Children() {
		if child.Uid() == t.Uid() {
			loc += i + 1
			break
		}
	}

	return
}

func (t *Table) Path() string {
	return t.parent.Path() + "/" + t.uid.String()
}
