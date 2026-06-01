package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp/table"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

type QueryTableInput struct {
	Queries   []mcp.GenericOneOf[*QueryTableUnion, TableApplicableTool] `json:"queries" jsonschema:"description=The list of queries to execute on the given table uids."`
	ShowDiff  bool                                                      `json:"show_diff,omitempty" jsonschema:"description=Whether to show a diff of the changes made; default is false,default=false"`
	HideTypes bool                                                      `json:"hide_types,omitempty" jsonschema:"description=Hide the #+TYPES specifier of the table (if one is available at all). This could be useful to ignore on repeated query call that do not require this information."`
	Path      string                                                    `json:"path,omitempty" jsonschema:"description=The path to the org file containing the tables to query. If not provided, the default path will be used."`
}

type TableApplyResult struct {
	rows  []table.TableRow
	types string
	raw   string
	err   error
}

type QueryTableUnion struct {
	tag string

	Sql    QueryTableSql
	Simple QueryTableSimple
}

func NewQueryTableUnion[T QueryTableSimple | QueryTableSql](t T) *QueryTableUnion {
	switch any(t).(type) {
	case QueryTableSimple:
		return &QueryTableUnion{
			tag:    "simple",
			Simple: any(t).(QueryTableSimple),
		}
	case QueryTableSql:
		return &QueryTableUnion{
			tag: "sql",
			Sql: any(t).(QueryTableSql),
		}
	}

	panic(fmt.Sprintf("unsupported type for QueryTableUnion: %s", reflect.TypeOf(t)))
}

func (q *QueryTableUnion) Tag() string {
	return q.tag
}

func (q *QueryTableUnion) Value() TableApplicableTool {
	switch q.tag {
	case "sql":
		return &q.Sql
	case "simple":
		return &q.Simple
	default:
		panic(fmt.Sprintf("invalid tag for QueryTableEntry: %s", q.tag))
	}
}

func (q *QueryTableUnion) FromJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch raw["method"] {
	case "sql":
		var sql QueryTableSql
		if err := json.Unmarshal(data, &sql); err != nil {
			return err
		}
		q.tag = "sql"
		q.Sql = sql
	case "simple":
		var simple QueryTableSimple
		if err := json.Unmarshal(data, &simple); err != nil {
			return err
		}
		q.tag = "simple"
		q.Simple = simple
	default:
		return fmt.Errorf("invalid method: %s", raw["method"])
	}

	return nil
}

type QueryTableSql struct {
	Uid    Uid    `json:"uid" jsonschema:"description=The uid of the table to query."`
	Method string `json:"method" jsonschema:"description=The method to use for the query.,enum=sql"`
	Query  string `json:"query" jsonschema:"description=The SQL query to execute on the given table uid."`
}

func (q *QueryTableSql) Apply(ctx context.Context) (res TableApplyResult) {
	var t *table.Table
	t, res.err = getTable(ctx, q.Uid)
	if res.err != nil {
		return
	}

	result, err := t.Query(q.Query)
	if err != nil {
		res.err = err
		return
	}

	res.raw = result

	return
}

type QueryTableSimple struct {
	Uid     Uid                    `json:"uid" jsonschema:"description=The uid of the table to query."`
	Method  string                 `json:"method" jsonschema:"description=The method to use for the query.,enum=simple"`
	Range   table.TableRange       `json:"range"` // TODO: support column selection as well
	Columns []table.ColumnSelector `json:"columns,omitempty" jsonschema:"description=Optional list of column selectors. Supports 0-based indices ('$0'), header names ('Name' or '${Name}'), and inclusive ranges ('$1:$3'). If omitted, returns all columns."`
}

func (q *QueryTableSimple) Apply(ctx context.Context) (res TableApplyResult) {
	var t *table.Table
	t, res.err = getTable(ctx, q.Uid)
	if res.err != nil {
		return
	}

	t, err := t.ApplyColumnSelectors(q.Columns)
	if err != nil {
		res.err = err
		return
	}

	rows, err := t.GetRange(q.Range)
	if err != nil {
		res.err = err
		return
	}

	res.rows = rows
	res.types = t.GetTypes()

	return
}

var QueryTableTool = mcp.GenericTool[QueryTableInput]{
	Name:        "query_table",
	Description: "This tool can query table in detail, with two methods for querying: SQL, and simple. For quick query that don't require complex queries simple is by far the prefered method to query. If calculation have to be run or more complex behaviour is needed the SQL method is the right choice. The SQL backend is sqlite so queries should keep it's dialect of SQL in mind.",
	Callback: func(ctx context.Context, input QueryTableInput, options mcp.FuncOptions) (resp []any, err error) {
		var path string
		if input.Path == "" {
			path = options.DefaultPath
		} else {
			path = input.Path
		}

		orgFile, err := mcp.LoadOrgFile(ctx, path)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, "orgfile", &orgFile)

		for _, q := range input.Queries {
			res := q.Value.Value().Apply(ctx)
			if res.err != nil {
				resp = append(resp, res.err)
			}

			if res.raw != "" {
				if !input.HideTypes && res.types != "" {
					resp = append(resp, "#+TYPE: "+res.types)
				}

				resp = append(resp, res.raw)
			} else if res.rows != nil {
				builder := strings.Builder{}

				if !input.HideTypes && res.types != "" {
					fmt.Fprintf(os.Stderr, "hide_types: %v\n", input.HideTypes)
					builder.WriteString("#+TYPE: " + res.types + "\n")
				}

				for _, row := range res.rows {
					builder.WriteString(row.String() + "\n")
				}

				resp = append(resp, builder.String())
			}
		}

		diff, err := mcp.WriteOrgFileToDisk(ctx, orgFile, path)
		if input.ShowDiff {
			resp = append(resp, diff)
		}

		return
	},
}
