package table

import (
	"fmt"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

var (
	ROW_OUT_OF_RANGE = "Index out of range of array length, to add a row use the `add` method. Table length: %d, got: %d"
)

func (t *Table) AddRow(idx int, row TableRow) {
	cr := t.ContentRows(false)

	if idx >= t.Len(false) {
		t.rows = append(t.rows, row)

		return
	}

	if idx < -t.Len(false) {
		idx = 0
	}

	// Bound the idx to at most the length of the slice -1
	boundedIdx := min(idx, len(cr))
	// make sure to modulo the negative indices to a positive from 0 to len(t.rows)
	// normal % does keep negatives which we do not want
	modIdx := ((boundedIdx)%(len(cr)+1) + (len(cr) + 1)) % (len(cr) + 1)


	newRows := []TableRow{}
	counter := 0
	for idx, r := range t.rows {
		if !r.HasContent() || idx == 0 {
			newRows = append(newRows, r)
			continue
		}

		if counter == modIdx {
			newRows = append(newRows, row)
		}

		counter += 1
		newRows = append(newRows, r)
	}

	if counter == modIdx {
		newRows = append(newRows, row)
	}

	t.rows = newRows
}

func (t *Table) UpdateRow(idx int, row TableRow) error {
	cr := t.ContentRows(false)

	if len(cr) <= idx || -len(cr) > idx {
		return fmt.Errorf(ROW_OUT_OF_RANGE, len(cr), idx)
	}


	// make sure to modulo the negative indices to a positive from 0 to len(t.rows)
	// normal % does keep negatives which we do not want
	modIdx := ((idx)%len(cr) + len(cr)) % len(cr)

	cr[modIdx].setCells(row.Items())

	return nil
}

func (t *Table) RemoveRow(idx int) error {
	cr := t.ContentRows(false)

	if len(cr) <= idx || -len(cr) > idx {
		return fmt.Errorf(ROW_OUT_OF_RANGE, len(cr), idx)
	}

	newRows := []TableRow{t.header}
	ptr := 0

	// make sure to modulo the negative indices to a positive from 0 to len(t.rows)
	// normal % does keep negatives which we do not want
	modIdx := ((idx)%len(cr) + len(cr)) % len(cr)

	for _, r := range t.rows[1:] {
		if !r.HasContent() {
			newRows = append(newRows, r)
			continue
		}

		if ptr != modIdx {
			newRows = append(newRows, r)
		}

		ptr += 1
	}

	t.rows = newRows

	return nil
}

func (t *Table) AddColumn(idx int, name string, d string) error {

	if idx < 0 {
		return fmt.Errorf(COLUMN_OUT_OF_RANGE, t, t.Uid(), len(t.GetHeader().Items()), idx)
	}

	t.columns += 1

	if idx >= t.columns {
		t.header.setCells(append(t.header.Items(), name))

		for _, r := range t.ContentRows(false) {
			r.setCells(append(r.Items(), d))
		}

		return nil
	}

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

	return nil
}

func (t *Table) UpdateColumn(idx int, name string) (string, error) {
	if idx < 0 {
		return "", fmt.Errorf(COLUMN_OUT_OF_RANGE, t, t.Uid(), len(t.GetHeader().Items()), idx)
	}

	headerItems := t.header.Items()

	if t.columns <= idx && len(headerItems) <= idx {
		return "", fmt.Errorf(COLUMN_OUT_OF_RANGE, t, t.Uid(), t.columns, idx)
	}

	oldName := headerItems[idx]
	headerItems[idx] = name

	t.header.setCells(headerItems)

	return oldName, nil
}

func (t *Table) RemoveColumn(idx int) (string, error) {
	if idx < 0 {
		return "", fmt.Errorf(COLUMN_OUT_OF_RANGE, t, t.Uid(), len(t.GetHeader().Items()), idx)
	}

	headerItems := t.header.Items()

	if t.columns <= idx && len(headerItems) <= idx {
		return "", fmt.Errorf(COLUMN_OUT_OF_RANGE, t, t.Uid(), t.columns, idx)
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
