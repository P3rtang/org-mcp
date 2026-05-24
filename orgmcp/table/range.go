package table

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/result"
)

var InvalidTableRange error = errors.New("invalid table range")

type TableRange struct {
	begin option.Option[int]
	end   option.Option[int]
}

func (t TableRange) GetSchema() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "The range of rows to return from the given table uid, in Python slice notation (e.g. '[1:5]' to return rows 1 through 4). This includes negative ranges like [-5:-1].",
	}
}

func NewTableRange(str string) result.Result[TableRange] {
	return result.TryFunc(func() (TableRange, error) { return parsePythonRange(str) })
}

func (t *TableRange) UnmarshalText(text []byte) error {
	r, err := parsePythonRange(string(text))
	t.begin = r.begin
	t.end = r.end

	return err
}

func (t *Table) GetRange(rg TableRange) ([]TableRow, error) {
	begin := rg.begin.UnwrapOr(0)
	end := rg.end.UnwrapOr(len(t.rows))

	filteredRows := []TableRow{}

	if len(t.rows) == 0 {
		return filteredRows, nil
	}

	for _, row := range t.rows[1:] {
		if !row.HasContent() {
			continue
		}

		filteredRows = append(filteredRows, row)
	}

	if begin < 0 {
		begin = len(filteredRows) + begin
	}

	if end < 0 {
		end = len(filteredRows) + end
	}

	if begin > end {
		slices.Reverse(filteredRows)
		begin, end = len(filteredRows)-begin, len(filteredRows)-end
	}

	if len(filteredRows) < begin {
		return filteredRows, fmt.Errorf("Start of table bounds out of range, table length: %d, got: %d", len(filteredRows), begin)
	}

	rows := []TableRow{t.rows[0]}
	rows = append(rows, filteredRows[begin:min(end, len(filteredRows))]...)

	return rows, nil
}

func parsePythonRange(str string) (r TableRange, err error) {
	if str == "" {
		err = InvalidTableRange
		return
	}

	if !strings.Contains(str, ":") {
		err = InvalidTableRange
		return
	}

	str = strings.TrimPrefix(str, "[")
	str = strings.TrimSuffix(str, "]")

	parts := strings.Split(str, ":")

	if len(parts) != 2 {
		err = InvalidTableRange
		return
	}

	if parts[0] != "" {
		var begin int
		_, err = fmt.Sscanf(parts[0], "%d", &begin)
		if err != nil {
			return
		}
		r.begin = option.Some(begin)
	} else {
		r.begin = option.None[int]()
	}

	if parts[1] != "" {
		var end int
		_, err = fmt.Sscanf(parts[1], "%d", &end)
		if err != nil {
			return
		}
		r.end = option.Some(end)
	} else {
		r.end = option.None[int]()
	}

	return
}
