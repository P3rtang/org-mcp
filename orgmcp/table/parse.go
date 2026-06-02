package table

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/reader"
)

var (
	ErrNotATable = errors.New("Line could not be parsed as a table.")
)

func isTableRow(line string) bool {
	return ((strings.HasPrefix(line, "| ") &&
		strings.HasSuffix(line, "|")) ||
		strings.HasPrefix(line, "|--")) ||
		strings.HasPrefix(line, "#+NAME:") ||
		strings.HasPrefix(line, "#+TYPE:")
}

// TODO: this probably has to be generalized to a section parser
func parseNameMetadata(line string) (uid Uid, ok bool) {
	prefix := "#+NAME:"
	line = strings.TrimSpace(line)

	ok = strings.HasPrefix(line, prefix)

	if ok {
		uidString := strings.TrimSpace(line[len(prefix):])
		uid = NewUid(uidString)
	}

	return
}

func parseTypeMetadata(line string) (types string, ok bool) {
	prefix := "#+TYPE:"

	line = strings.TrimSpace(line)

	ok = strings.HasPrefix(line, prefix)

	if ok {
		return strings.TrimSpace(line[len(prefix):]), ok
	}

	return
}

func NewTableFromReader(r *reader.PeekReader) (t Table, err error) {
	var bytes []byte
	bytes, err = r.PeekBytes('\n')
	if err != nil {
		return
	}

	if name, ok := parseNameMetadata(string(bytes)); ok {
		r.Discard()
		bytes, err = r.PeekBytes('\n')

		if err != nil {
			return
		}

		t.uid = name
	}

	if !isTableRow(strings.TrimSpace(string(bytes))) {
		err = ErrNotATable

		return
	}

	for ; err == nil && isTableRow(strings.TrimSpace(string(bytes))); bytes, err = r.PeekBytes('\n') {
		r.Discard()

		line := strings.TrimSpace(string(bytes))

		fmt.Fprintf(os.Stderr, "Parsing table line: %s\n", line)

		if strings.HasPrefix(line, "| ") {
			items := strings.Split(line[1:len(line)-1], "|")

			for i, item := range items {
				items[i] = strings.TrimSpace(item)
			}

			if t.columns == 0 {
				t.columns = len(items)
			}

			t.rows = append(t.rows, &ContentRow{cells: items[0:min(len(items), t.columns)]})
		} else if strings.HasPrefix(line, "|--") {
			t.rows = append(t.rows, &DividerRow{})
		} else if meta, ok := parseNameMetadata(line); ok {
			t.uid = meta
		} else if types, ok := parseTypeMetadata(line); ok {
			t.types = types
		} else {
			err = ErrNotATable
		}
	}

	if t.uid.String() == "" {
		t.uid = NewUid(fmt.Sprintf("table_%d", rand.Intn(100000000)))
	}

	if errors.Is(err, ErrNotATable) || errors.Is(err, io.EOF) {
		err = nil
	}

	t.header = t.rows[0]

	return
}
