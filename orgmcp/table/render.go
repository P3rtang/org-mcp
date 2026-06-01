package table

import (
	"fmt"
	"strings"
)

func (t *Table) Render(
	builder *strings.Builder,
	// Depth is ignored since a table is always a leaf node
	_ int,
) {
	indent := t.parent.ChildIndentLevel()
	indentStr := strings.Repeat(" ", indent)
	colWidth := make([]int, t.columns)

	for _, row := range t.rows {
		for i, w := range row.ColWidths() {
			if i >= len(colWidth) {
				break
			}

			colWidth[i] = max(colWidth[i], w)
		}
	}

	for _, row := range t.rows {
		builder.WriteString(indentStr)
		row.Render(builder, colWidth)
		builder.WriteRune('\n')
	}

	fmt.Fprintf(builder, "%s#+NAME: %s\n", indentStr, t.uid)
	fmt.Fprintf(builder, "%s#+TYPE: %s\n", indentStr, t.types)
}

func (t *Table) RenderMarkdown(builder *strings.Builder, depth int) {}

// Ignore the length argument since we want to inform with a consistent message regardless of the length of the preview.
func (t *Table) Preview(int) string {
	builder := strings.Builder{}
	t.rows[0].Render(&builder, t.rows[0].ColWidths())

	return builder.String()
}

func (tr *ContentRow) Render(b *strings.Builder, widths []int) {
	items := tr.Items()

	for i, w := range widths {
		var item string
		if len(items) > i {
			item = items[i]
		}

		fmt.Fprintf(b, "| %-*s ", w, item)
	}

	b.WriteRune('|')
}

func (tr *DividerRow) Render(b *strings.Builder, widths []int) {
	b.WriteRune('|')

	for i, width := range widths {
		fmt.Fprintf(b, "%s", strings.Repeat("-", width+2))

		if i < len(widths)-1 {
			b.WriteRune('+')
		}
	}

	b.WriteRune('|')
}
