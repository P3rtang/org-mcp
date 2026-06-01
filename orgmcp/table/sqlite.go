package table

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/itertools"
)

func (t *Table) createTableQuery(uid Uid) string {
	b := strings.Builder{}

	fmt.Fprintf(&b, "CREATE TABLE \"%s\" ", uid)

	if len(t.rows) == 0 {
		return ""
	}

	items := itertools.Collect(itertools.Map(itertools.FromSlice(t.rows[0].Items()), func(i string) string {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(i, "'", "\\'"))
	}))

	fmt.Fprintf(&b, "(%s)", strings.Join(items, ","))

	return b.String()
}

func (t *Table) populateTableQuery(uid Uid) string {
	b := strings.Builder{}

	fmt.Fprintf(&b, "INSERT INTO \"%s\" VALUES ", uid)

	if len(t.rows) <= 1 {
		return ""
	}

	for i, row := range t.rows[1:] {
		items := itertools.Collect(itertools.Map(itertools.FromSlice(row.Items()), func(i string) string {
			return fmt.Sprintf("'%s'", strings.ReplaceAll(i, "'", "\\'"))
		}))

		if len(row.Items()) == 0 {
			continue
		}

		fmt.Fprintf(&b, "(%s)", strings.Join(items, ","))

		if i != len(t.rows)-2 {
			b.WriteString(",")
		}
	}

	return b.String()
}

func (t *Table) Query(q string) (string, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return "", err
	}

	// Add the table for both the full uid + the short form uid (just table name)
	// This allows agent to use both without failing and needing to rerun
	_, err = db.Exec(t.createTableQuery(t.uid))
	_, err = db.Exec(t.createTableQuery(t.Uid()))
	_, err = db.Exec(t.populateTableQuery(t.uid))
	_, err = db.Exec(t.populateTableQuery(t.Uid()))

	// panic(fmt.Sprintf("----------------\n%s\n%s\n---------------------\n", t.createTableQuery(), t.populateTableQuery()))

	if err != nil {
		return "", err
	}

	rows, err := db.Query(q)
	if err != nil {
		return "", err
	}

	csvB := strings.Builder{}
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}

	csvB.WriteString(strings.Join(columns, ","))
	csvB.WriteRune('\n')

	for rows.Next() {
		values := make([]string, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)

		csvB.WriteString(strings.Join(values, ","))
		csvB.WriteRune('\n')
	}

	fmt.Println(csvB.String())

	return csvB.String(), nil
}
