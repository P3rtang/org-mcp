package table

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/p3rtang/org-mcp/utils/slice"
)

const (
	FROM_COLUMN_NOT_FOUND = "Did not find the column associated with the start of the range. Open ended ranges are not allowed with named columns."
	TO_COLUMN_NOT_FOUND   = "Did not find the column associated with the end of the range. Open ended ranges are not allowed with named columns."
	BACKWARDS_INDECES     = "Column ranges are not allowed to be backwards, the starting index has to be lower than the ending index."
	COLUMN_NOT_FOUND      = "The specified column name could not be found. Got: %s, Possible values: %s"
	COLUMN_OUT_OF_RANGE   = "The specified column index was out of range. Columns: %d, got: %d."
	INVALID_SELECTOR      = "Invalid pattern for the column selector, use either numeric columns `$1` or named columns `${name}`. Got: %s"
	TOO_MANY_COLONS       = "A selector string cannot contain more than one : outside of ${} column names."
	CANNOT_COMBINE        = "A combination of numeric and named selectors are not allowed within a single range. Use either `$1:$2` or `${col1}:${col2}`"
)

type ColumnRangeNamed struct {
	from string
	to   string
}

func (c *ColumnRangeNamed) cols(t *Table) (cols []int, err error) {
	names := t.header.Items()

	for idx, name := range names {
		if name == c.to {
			if len(cols) == 0 {
				return []int{}, fmt.Errorf(FROM_COLUMN_NOT_FOUND)
			}

			cols = append(cols, idx)

			// Happy path both from and to were found
			return
		}

		if name == c.from || len(cols) > 0 {
			cols = append(cols, idx)
		}
	}

	return []int{}, fmt.Errorf(TO_COLUMN_NOT_FOUND)
}

type ColumnRangeIndexed struct {
	from int
	to   int
}

func (c *ColumnRangeIndexed) cols(t *Table) (cols []int, err error) {
	if c.from > c.to {
		err = fmt.Errorf(BACKWARDS_INDECES)
	}

	for i := c.from; i <= c.to; i += 1 {
		cols = append(cols, i)
	}

	return
}

type ColumnNamed struct {
	name string
}

func (c *ColumnNamed) cols(t *Table) (cols []int, err error) {
	for idx, name := range t.header.Items() {
		if name == c.name {
			cols = append(cols, idx)

			return
		}
	}

	return cols, fmt.Errorf(COLUMN_NOT_FOUND, c.name, slice.Joins(t.header.Items(), ","))
}

type ColumnIndex struct {
	idx int
}

func (c *ColumnIndex) cols(t *Table) (cols []int, err error) {
	if c.idx > t.columns {
		err = fmt.Errorf(COLUMN_OUT_OF_RANGE, t.columns, c.idx)
		return
	}

	cols = append(cols, c.idx)
	return
}

type ToColumns interface {
	cols(t *Table) ([]int, error)
}

type ColumnSelector struct {
	cols ToColumns
}

type SelectorPart struct {
	name string
	num  int

	isNumeric bool
}

func newSelectorPart(str string) (SelectorPart, error) {
	err := fmt.Errorf(INVALID_SELECTOR, str)

	if !strings.HasPrefix(str, "$") || len(str) <= 1 {
		return SelectorPart{}, err
	}

	// For any part still containing : strip it and the leading $ sign
	str = strings.TrimSuffix(str[1:], ":")

	if strings.HasPrefix(str, "{") {
		if !strings.HasSuffix(str, "}") {
			return SelectorPart{}, err
		}

		return SelectorPart{
			name:      str[1 : len(str)-1],
			isNumeric: false,
		}, nil
	}

	if num, err := strconv.Atoi(str); err == nil {
		return SelectorPart{
			num:       num,
			isNumeric: true,
		}, nil
	}

	return SelectorPart{}, err
}

func newToColumns(bytes []byte) (ToColumns, error) {
	parts := []string{"", ""}
	partIdx := 0
	openBracket := false

	for _, byte := range bytes {
		switch byte {
		case ':':
			if !openBracket {
				if partIdx == 1 {
					return nil, fmt.Errorf(INVALID_SELECTOR, "more than 1 colons for range format.")
				}

				partIdx = 1
				break
			}
			fallthrough
		case '{':
			fallthrough
		case '}':
			openBracket = byte == '{'
			fallthrough
		default:
			parts[partIdx] += string(byte)
		}
	}

	parts = slice.Filter(parts, func(s string) bool { return s != "" })

	selectorParts, err := slice.TryMap(parts, func(s string) (SelectorPart, error) {
		return newSelectorPart(s)
	})

	if err != nil {
		return nil, err
	}

	if len(selectorParts) == 1 {
		if selectorParts[0].isNumeric {
			return &ColumnIndex{idx: selectorParts[0].num}, nil
		} else {
			return &ColumnNamed{name: selectorParts[0].name}, nil
		}
	}

	if selectorParts[0].isNumeric != selectorParts[1].isNumeric {
		return nil, fmt.Errorf(CANNOT_COMBINE)
	}

	if selectorParts[0].isNumeric {
		return &ColumnRangeIndexed{from: selectorParts[0].num, to: selectorParts[1].num}, nil
	} else {
		return &ColumnRangeNamed{from: selectorParts[0].name, to: selectorParts[1].name}, nil
	}
}

func (cs *ColumnSelector) UnmarshalText(bytes []byte) (err error) {
	c, err := newToColumns(bytes)

	if err != nil {
		fmt.Printf("\n%s\n", err.Error())
		return err
	}

	cs.cols = c

	return
}
