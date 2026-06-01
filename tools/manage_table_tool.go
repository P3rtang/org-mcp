package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/orgmcp/table"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/option"
)

var (
	COLUMN_ADDED   = "Column with name %s added. New header row: %s."
	COLUMN_UPDATED = "Column %s renamed to %s. New header row: %s."
	COLUMN_REMOVED = "Column %s removed. New header row: %s."
)

func getTable(ctx context.Context, uid Uid) (*table.Table, error) {
	of, ok := ctx.Value("orgfile").(*orgmcp.OrgFile)
	if !ok {
		return nil, fmt.Errorf("Could not find the org file in testing context")
	}

	var t *table.Table
	if t, ok = option.Cast[Render, *table.Table](of.GetUid(uid)).Split(); !ok {
		return nil, fmt.Errorf("Could not find table with uid `%s`.", uid)
	}

	return t, nil
}

type TaggedUnion = mcp.GenericOneOf[*ManageTableInputUnion, TableApplicableTool]

type ManageTableInput struct {
	Items        []TaggedUnion `json:"items" jsonschema:"description=List of table operations to perform. Multiple operations can be performed in a single call."`
	Path         string        `json:"path,omitempty" jsonschema:"description=The file path to the Org file to modify. It will target the ./.tasks.org by default and you don't have to pass this in unless you want to target a different file.,required=false"`
	ShowDiff     bool          `json:"show_diff,omitempty" jsonschema:"description=Whether to return the diff of changes made to the file. Can be used to inform you and the user of what changed.,required=false"`
	ShowAffected *bool         `json:"show_affected,omitempty" jsonschema:"description=Whether to include the affected items in the response. This will include direct neighbors of the affected item as well.,default=true,required=false"`
}

type ManageTableInputUnion struct {
	tag string

	Add       AddTableRow
	Update    UpdateTableRow
	Remove    RemoveTableRow
	AddCol    AddTableColumn
	UpdateCol UpdateTableColumn
	RemoveCol RemoveTableColumn
}

func NewManageTableInputUnion[T AddTableRow | UpdateTableRow | RemoveTableRow | AddTableColumn | UpdateTableColumn | RemoveTableColumn](input T) *ManageTableInputUnion {
	switch any(input).(type) {
	case AddTableRow:
		return &ManageTableInputUnion{
			tag: "add",
			Add: any(input).(AddTableRow),
		}
	case UpdateTableRow:
		return &ManageTableInputUnion{
			tag:    "update",
			Update: any(input).(UpdateTableRow),
		}
	case RemoveTableRow:
		return &ManageTableInputUnion{
			tag:    "remove",
			Remove: any(input).(RemoveTableRow),
		}
	case AddTableColumn:
		return &ManageTableInputUnion{
			tag:    "add_col",
			AddCol: any(input).(AddTableColumn),
		}
	case UpdateTableColumn:
		return &ManageTableInputUnion{
			tag:       "update_col",
			UpdateCol: any(input).(UpdateTableColumn),
		}
	case RemoveTableColumn:
		return &ManageTableInputUnion{
			tag:       "remove_col",
			RemoveCol: any(input).(RemoveTableColumn),
		}
	default:
		panic(fmt.Sprintf("unsupported type for ManageTableInputUnion: %s", fmt.Sprintf("%T", input)))
	}
}

func (u ManageTableInputUnion) Uid() Uid {
	switch u.tag {
	case "add":
		return u.Add.Uid
	case "update":
		return u.Update.Uid
	case "remove":
		return u.Remove.Uid
	case "add_col":
		return u.AddCol.Uid
	case "update_col":
		return u.UpdateCol.Uid
	case "remove_col":
		return u.RemoveCol.Uid
	default:
		panic("unreachable")
	}
}

func (u ManageTableInputUnion) Apply(ctx context.Context) (res TableApplyResult) {
	switch u.tag {
	case "add":
		return u.Add.Apply(ctx)
	case "update":
		return u.Update.Apply(ctx)
	case "remove":
		return u.Remove.Apply(ctx)
	case "add_col":
		return u.AddCol.Apply(ctx)
	case "update_col":
		return u.UpdateCol.Apply(ctx)
	case "remove_col":
		return u.RemoveCol.Apply(ctx)
	default:
		panic("unreachable")
	}
}

func (u ManageTableInputUnion) Tag() string {
	return u.tag
}

func (u ManageTableInputUnion) Value() TableApplicableTool {
	switch u.tag {
	case "add":
		return &u.Add
	case "update":
		return &u.Update
	case "remove":
		return &u.Remove
	case "add_col":
		return &u.AddCol
	case "update_col":
		return &u.UpdateCol
	case "remove_col":
		return &u.RemoveCol
	default:
		return nil
	}
}

func (u *ManageTableInputUnion) FromJSON(data []byte) error {
	var temp struct {
		Method string `json:"method"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	switch temp.Method {
	case "add":
		u.tag = "add"
		return json.Unmarshal(data, &u.Add)
	case "update":
		u.tag = "update"
		return json.Unmarshal(data, &u.Update)
	case "remove":
		u.tag = "remove"
		return json.Unmarshal(data, &u.Remove)
	case "add_col":
		u.tag = "add_col"
		return json.Unmarshal(data, &u.AddCol)
	case "update_col":
		u.tag = "update_col"
		return json.Unmarshal(data, &u.UpdateCol)
	case "remove_col":
		u.tag = "remove_col"
		return json.Unmarshal(data, &u.RemoveCol)
	default:
		return fmt.Errorf("invalid method for table operation: %s", temp.Method)
	}
}

type AddTableRow struct {
	Method string   `json:"method" jsonschema:"enum=add"`
	Uid    Uid      `json:"uid" jsonschema:"description=The uid of the table to modify."`
	Row    int      `json:"row" jsonschema:"description=The row number you want this new element to sit at. Keep in mind this will bump row numbers for every row behind."`
	Cells  []string `json:"cells" jsonschema:"description=The cell values for the new row."`
}

type UpdateTableRow struct {
	Method string   `json:"method" jsonschema:"enum=update"`
	Uid    Uid      `json:"uid" jsonschema:"description=The uid of the table to modify."`
	Row    int      `json:"row" jsonschema:"description=The row number of the table element you want to edit."`
	Cells  []string `json:"cells" jsonschema:"description=The new cell values for the row."`
}

type RemoveTableRow struct {
	Method string `json:"method" jsonschema:"enum=remove"`
	Uid    Uid    `json:"uid" jsonschema:"description=The uid of the table to modify."`
	Row    int    `json:"row" jsonschema:"description=The row number of the table element you want to remove. Keep in mind this will bump row numbers for every row behind."`
}

type AddTableColumn struct {
	Method  string `json:"method" jsonschema:"enum=add_col"`
	Uid     Uid    `json:"uid" jsonschema:"description=The uid of the table to modify."`
	Col     int    `json:"col" jsonschema:"description=The column number to add the new column at."`
	Name    string `json:"name" jsonschema:"description=The header name for the new column."`
	Default string `json:"default,omitempty" jsonschema:"description=The default value for the new column. This is only used in this tool call and not for future rows."`
}

type UpdateTableColumn struct {
	Method string `json:"method" jsonschema:"enum=update_col"`
	Uid    Uid    `json:"uid" jsonschema:"description=The uid of the table to modify."`
	Col    int    `json:"col" jsonschema:"description=The column number to update."`
	Name   string `json:"name" jsonschema:"description=The new header name for the column."`
}

type RemoveTableColumn struct {
	Method string `json:"method" jsonschema:"enum=remove_col"`
	Uid    Uid    `json:"uid" jsonschema:"description=The uid of the table to modify."`
	Col    int    `json:"col" jsonschema:"description=The column number to remove."`
}

func (a *AddTableRow) Apply(ctx context.Context) (res TableApplyResult) {
	var t *table.Table

	t, res.err = getTable(ctx, a.Uid)
	if res.err != nil {
		return
	}

	row := table.NewContentRow(a.Cells)
	t.AddRow(a.Row, &row)

	res.rows, res.err = t.GetRange(table.NewTableRangeFull(max(a.Row-1, 0), a.Row+2))

	return
}

func (u *UpdateTableRow) Apply(ctx context.Context) (res TableApplyResult) {
	var t *table.Table

	t, res.err = getTable(ctx, u.Uid)
	if res.err != nil {
		return
	}

	row := table.NewContentRow(u.Cells)
	res.err = t.UpdateRow(u.Row, &row)
	if res.err != nil {
		return
	}

	res.rows, res.err = t.GetRange(table.NewTableRangeFull(max(u.Row-1, 0), u.Row+2))

	return
}

func (r *RemoveTableRow) Apply(ctx context.Context) (res TableApplyResult) {
	var t *table.Table

	t, res.err = getTable(ctx, r.Uid)
	if res.err != nil {
		return
	}

	res.err = t.RemoveRow(r.Row)
	if res.err != nil {
		return
	}

	res.rows, res.err = t.GetRange(table.NewTableRangeFull(max(r.Row-1, 0), r.Row+1))

	return
}

func (r *AddTableColumn) Apply(ctx context.Context) (res TableApplyResult) {
	var t *table.Table

	t, res.err = getTable(ctx, r.Uid)
	if res.err != nil {
		return
	}

	t.AddColumn(r.Col, r.Name, r.Default)

	res.raw = fmt.Sprintf(COLUMN_ADDED, r.Name, t.GetHeader())

	return
}

func (r *UpdateTableColumn) Apply(ctx context.Context) (res TableApplyResult) {
	var t *table.Table

	t, res.err = getTable(ctx, r.Uid)
	if res.err != nil {
		return
	}

	var u string
	u, res.err = t.UpdateColumn(r.Col, r.Name)

	res.raw = fmt.Sprintf(COLUMN_UPDATED, u, r.Name, t.GetHeader())

	return
}

func (r *RemoveTableColumn) Apply(ctx context.Context) (res TableApplyResult) {
	var t *table.Table

	t, res.err = getTable(ctx, r.Uid)
	if res.err != nil {
		return
	}

	var rem string
	rem, res.err = t.RemoveColumn(r.Col)
	if res.err != nil {
		return
	}

	res.raw = fmt.Sprintf(COLUMN_REMOVED, rem, t.GetHeader())

	return
}

var ManageTableTool = mcp.GenericTool[ManageTableInput]{
	Name:        "manage_table",
	Description: "Add, update or remove rows in a table within an Org file.",
	Callback: func(ctx context.Context, input ManageTableInput, options mcp.FuncOptions) (resp []any, err error) {
		var path string
		if input.Path == "" {
			path = options.DefaultPath
		} else {
			path = input.Path
		}

		orgFile, err := mcp.LoadOrgFile(ctx, path)
		ctx = orgFile.OrgFileToContext(ctx)
		if err != nil {
			return
		}

		for _, i := range input.Items {
			res := i.Value.Apply(ctx)

			if res.err != nil {
				resp = append(resp, res.err)
			} else if res.raw != "" {
				resp = append(resp, res.raw)
			} else if res.rows != nil {
				b := strings.Builder{}
				l := len(res.rows[0].Items())

				for _, row := range res.rows {
					if row == nil {
						continue
					}

					items := make([]string, l)
					copy(items, row.Items())

					fmt.Fprintf(&b, "%s\n", strings.Join(items, ","))
				}
				resp = append(resp, b.String())
			}
		}

		diff, err := mcp.WriteOrgFileToDisk(ctx, orgFile, path)

		if err != nil {
			return nil, err
		}

		if input.ShowDiff {
			resp = append(resp, diff)
		}

		return
	},
}

// intoManageTableOneOfArray helper for tests
func intoManageTableOneOfArray[T AddTableRow | UpdateTableRow | RemoveTableRow](t ...T) []mcp.GenericOneOf[*ManageTableInputUnion, TableApplicableTool] {
	entries := []mcp.GenericOneOf[*ManageTableInputUnion, TableApplicableTool]{}
	for _, input := range t {
		entries = append(entries, mcp.GenericOneOf[*ManageTableInputUnion, TableApplicableTool]{
			Value: NewManageTableInputUnion(input),
		})
	}
	return entries
}

func renderTableRows(rows []table.TableRow) string {
	return ""
}
