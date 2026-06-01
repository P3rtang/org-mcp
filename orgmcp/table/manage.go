package table

import (
	"fmt"
	"os"
)

var (
	ROW_OUT_OF_RANGE = "Index out of range of array length, to add a row use the `add` method. Table length: %d, got: %d"
)

func (t *Table) AddRow(idx int, row TableRow) {
	newRows := []TableRow{t.header}
	ptr := 0

	for _, r := range t.rows[1:] {
		newRows = append(newRows, r)
		if r.HasContent() {
			ptr += 1
		}

		if ptr == idx {
			newRows = append(newRows, row)
		}
	}

	t.rows = newRows
}

func (t *Table) UpdateRow(idx int, row TableRow) error {
	cr := t.ContentRows(false)

	if len(cr) <= idx {
		return fmt.Errorf(ROW_OUT_OF_RANGE, len(cr), idx)
	}

	cr[idx].setCells(row.Items())

	return nil
}

func (t *Table) RemoveRow(idx int) error {
	cr := t.ContentRows(false)

	if len(cr) <= idx {
		return fmt.Errorf(ROW_OUT_OF_RANGE, len(cr), idx)
	}

	newRows := []TableRow{t.header}
	ptr := 0

	for _, r := range t.rows[1:] {
		if !r.HasContent() {
			newRows = append(newRows, r)
			continue
		}

		if ptr != idx {
			newRows = append(newRows, r)
		}

		ptr += 1
	}

	t.rows = newRows

	return nil
}

func (t *Table) AddColumn(idx int, name string, d string) {
	t.columns += 1

	if idx >= t.columns {
		t.header.setCells(append(t.header.Items(), name))

		for _, r := range t.ContentRows(false) {
			r.setCells(append(r.Items(), d))
		}

		return
	}

	fmt.Fprintf(os.Stderr, "rows: %v\n", t.ContentRows(true))

	for i, r := range t.ContentRows(true) {
		colBefore := r.Items()[:idx]
		colAfter := r.Items()[idx:]

		var value string
		if i == 0 {
			value = name
		} else {
			value = d
		}

		newCells := make([]string, len(colBefore)+len(colAfter)+1)
		copy(newCells, colBefore)
		newCells[idx] = value
		copy(newCells[idx+1:], colAfter)

		r.setCells(newCells)
	}
}

func (t *Table) UpdateColumn(idx int, name string) (string, error) {
	headerItems := t.header.Items()

	if t.columns <= idx && len(headerItems) <= idx {
		return "", fmt.Errorf(COLUMN_OUT_OF_RANGE, t.columns, idx)
	}

	oldName := headerItems[idx]
	headerItems[idx] = name

	t.header.setCells(headerItems)

	return oldName, nil
}

func (t *Table) RemoveColumn(idx int) (string, error) {
	headerItems := t.header.Items()

	if t.columns <= idx && len(headerItems) <= idx {
		return "", fmt.Errorf(COLUMN_OUT_OF_RANGE, t.columns, idx)
	}

	for _, r := range t.ContentRows(true) {
		items := r.Items()

		if idx == len(items)-1 {
			r.setCells(items[:idx])

			continue
		}

		r.setCells(append(items[:idx], items[idx+1:]...))
	}

	t.columns -= 1
	return headerItems[idx], nil
}
