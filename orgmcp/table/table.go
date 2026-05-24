package table

import (
	"strings"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/slice"
)

type TableRow interface {
	ColWidths() []int
	Items() []string
	HasContent() bool
	Render(*strings.Builder, []int)
}

type Table struct {
	uid    Uid
	parent Render

	columns int
	rows    []TableRow
}

// Assert Table is a Render interface
var _ Render = (*Table)(nil)

type DividerRow struct{}

func (tr *DividerRow) ColWidths() []int {
	return []int{}
}

func (tr *DividerRow) Items() []string {
	return []string{}
}

func (tr *DividerRow) HasContent() bool {
	return false
}

type ContentRow struct {
	items []string
}

func (tr *ContentRow) ColWidths() (widths []int) {
	for _, item := range tr.items {
		widths = append(widths, len(item))
	}

	return
}

func (tr *ContentRow) Items() []string {
	return tr.items
}

func (tr *ContentRow) HasContent() bool {
	return slice.Any(tr.items, func(str string) bool {
		return str != ""
	})
}
