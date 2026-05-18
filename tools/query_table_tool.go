package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/orgmcp/table"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

type QueryTableInput struct {
	Queries  []mcp.GenericOneOf[*QueryTableUnion, ApplicableTool] `json:"queries" jsonschema:"description=The list of queries to execute on the given table uids."`
	ShowDiff bool                                                 `json:"show_diff,omitempty" jsonschema:"description=Whether to show a diff of the changes made; default is false,default=false"`
	Path     string                                               `json:"path,omitempty" jsonschema:"description=The path to the org file containing the tables to query. If not provided, the default path will be used."`
}

type TableApplyResult struct {
	rows []table.TableRow
	raw  string
	err  error
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

func (q *QueryTableUnion) Value() ApplicableTool {
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

func (q *QueryTableSql) Apply(ctx context.Context, of *orgmcp.OrgFile) (res TableApplyResult, err error) {
	var selected Render
	var ok bool

	if selected, ok = of.GetUid(q.Uid).Split(); !ok {
		err = fmt.Errorf("Uid %s not found in %s", q.Uid, of.Name())
		return
	}

	table, ok := selected.(*table.Table)

	result, err := table.Query(q.Query)
	if err != nil {
		err = fmt.Errorf("Error executing SQL query on table with UID %s: %v", q.Uid, err)
		return
	}

	res.raw = result

	return
}

type QueryTableSimple struct {
	Uid    Uid              `json:"uid" jsonschema:"description=The uid of the table to query."`
	Method string           `json:"method" jsonschema:"description=The method to use for the query.,enum=simple"`
	Range  table.TableRange `json:"range"` // TODO: support column selection as well
}

func (q *QueryTableSimple) Apply(ctx context.Context, of *orgmcp.OrgFile) (res TableApplyResult, err error) {
	var selected Render
	var ok bool

	if selected, ok = of.GetUid(q.Uid).Split(); !ok {
		err = fmt.Errorf("Uid %s not found in %s", q.Uid, of.Name())
		return
	}

	table, ok := selected.(*table.Table)
	if !ok {
		err = fmt.Errorf("Item with UID %s is not a table in %s", q.Uid, of.Name())
		return
	}

	rows, err := table.GetRange(q.Range)

	if err != nil {
		err = fmt.Errorf("Error executing simple query on table with UID %s: %v", q.Uid, err)
		return
	}

	res.rows = rows

	return
}

var QueryTableTool = mcp.GenericTool[QueryTableInput]{
	Name:        "query_table",
	Description: "Executes a SQL query on a table with the given uid and returns the result as a string.",
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

		for _, q := range input.Queries {
			res, err := q.Value.Value().Apply(ctx, &orgFile)
			if err != nil {
				return nil, err
			}

			if res.err != nil {
				return nil, res.err
			}

			if res.raw != "" {
				resp = append(resp, res.raw)
			} else if res.rows != nil {
				b := strings.Builder{}
				for _, row := range res.rows {
					if row == nil {
						continue
					}
					fmt.Fprintf(&b, "%s\n", strings.Join(row.Items(), ","))
				}
				resp = append(resp, b.String())
			}
		}

		diff, err := mcp.WriteOrgFileToDisk(ctx, orgFile, path)
		if input.ShowDiff {
			resp = append(resp, diff)
		}

		return
	},
}
