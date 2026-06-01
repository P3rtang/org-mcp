package table

import (
	"strings"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/slice"
)

func escapePipe(in []string) []string {
	return slice.Map(in, func(str string) string {
		if strings.HasSuffix(str, "|") {
			str = str[:len(str)-1] + "\\vert"
		}

		return strings.ReplaceAll(str, "|", "\\vert{}")
	})
}

type TableRow interface {
	ColWidths() []int
	Items() []string
	setCells([]string)
	HasContent() bool
	Render(*strings.Builder, []int)
	String() string
}

type Table struct {
	uid    Uid
	parent Render

	columns int
	header  TableRow
	rows    []TableRow
}

func (t *Table) ContentRows(includeHeader bool) (rows []TableRow) {
	for idx, r := range t.rows {
		if r.HasContent() {
			if !includeHeader && idx == 0 {
				continue
			}

			rows = append(rows, r)
		}
	}

	return
}

func (t *Table) GetHeader() TableRow {
	return t.header
}

// Assert Table is a Render interface
var _ Render = (*Table)(nil)

type DividerRow struct{}

func (tr *DividerRow) ColWidths() []int  { return []int{} }
func (tr *DividerRow) Items() []string   { return []string{} }
func (tr *DividerRow) setCells([]string) {}
func (tr *DividerRow) HasContent() bool  { return false }
func (tr *DividerRow) String() string    { return "<divider/>" }

type ContentRow struct {
	cells []string
}

func NewContentRow(cells []string) ContentRow {
	return ContentRow{cells}
}

func (tr *ContentRow) ColWidths() (widths []int) {
	for _, item := range tr.cells {
		widths = append(widths, len(item))
	}

	return
}

func (tr *ContentRow) Items() []string {
	return tr.cells
}

func (tr *ContentRow) HasContent() bool {
	return slice.Any(tr.cells, func(str string) bool {
		return str != ""
	})
}

func (tr *ContentRow) setCells(cells []string) {
	tr.cells = escapePipe(cells)
}

func (tr *ContentRow) String() string {
	builder := strings.Builder{}

	builder.WriteString(slice.Joins(slice.Map(tr.cells, func(str string) string {
		str = strings.ReplaceAll(strings.ReplaceAll(str, "\\vert{}", "|"), "\\vert", "|")
		if strings.ContainsRune(str, ',') {
			str = "\"" + str + "\""
		}

		return str
	}), ","))

	return builder.String()
}
