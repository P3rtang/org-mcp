package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	codeblock "github.com/p3rtang/org-mcp/orgmcp/code-block"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/slice"
)

type CodeBlockUnion = mcp.GenericOneOf[*ManageCodeBlockInputUnion, CodeBlockApplicableTool]

type CodeBlockApplicableTool interface {
	Apply(ctx context.Context) CodeBlockApplyResult
}

type CodeBlockApplyResult struct {
	affected []Render
	raw      string
	err      error
}

type ManageCodeBlockInputUnion struct {
	tag string

	add    AddCodeBlock
	update UpdateCodeBlock
	remove RemoveCodeBlock
}

func NewManageCodeBlockInputUnion[T AddCodeBlock | UpdateCodeBlock | RemoveCodeBlock](t T) *ManageCodeBlockInputUnion {
	switch any(t).(type) {
	case AddCodeBlock:
		return &ManageCodeBlockInputUnion{tag: "add", add: any(t).(AddCodeBlock)}
	case UpdateCodeBlock:
		return &ManageCodeBlockInputUnion{tag: "update", update: any(t).(UpdateCodeBlock)}
	case RemoveCodeBlock:
		return &ManageCodeBlockInputUnion{tag: "remove", remove: any(t).(RemoveCodeBlock)}
	default:
		panic("unreachable")
	}
}

func (mcb *ManageCodeBlockInputUnion) Tag() string { return mcb.tag }

func (mcb *ManageCodeBlockInputUnion) Value() CodeBlockApplicableTool {
	switch mcb.Tag() {
	case "add":
		return mcb.add
	case "update":
		return mcb.update
	case "remove":
		return mcb.remove
	default:
		panic("unreachable")
	}
}

func (mcb *ManageCodeBlockInputUnion) FromJSON(data []byte) error {
	var temp struct {
		Method string `json:"method"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	switch temp.Method {
	case "add":
		mcb.tag = "add"
		return json.Unmarshal(data, &mcb.add)
	case "update":
		mcb.tag = "update"
		return json.Unmarshal(data, &mcb.update)
	case "remove":
		mcb.tag = "remove"
		return json.Unmarshal(data, &mcb.remove)
	default:
		return fmt.Errorf("invalid method for table operation: %s", temp.Method)
	}
}

type ManageCodeBlockInput struct {
	Items        []CodeBlockUnion `json:"items"`
	Path         string           `json:"path,omitempty" jsonschema:"description=The file path to the Org file to modify. It will target the ./.tasks.org by default and you don't have to pass this in unless you want to target a different file.,required=false"`
	ShowDiff     bool             `json:"show_diff,omitempty" jsonschema:"description=The diff format for the changes made to the org file. Can be used to track side effects of certain tool calls and to debug.,required=false"`
	HideAffected bool             `json:"hide_affected,omitempty" jsonschema:"description=Flag to hide affected items from the response. This is feedback you want to leave on for most cases.,default=false"`
	Columns      []*orgmcp.Column `json:"columns,omitempty" jsonschema:"description=List of columns to include in the output. If not specified defaults to [UID | PREVIEW]."`
}

type AddCodeBlock struct {
	Method string `json:"method" jsonschema:"enum=add,required=true"`
	Parent Uid    `json:"parent" jsonschema:"description=UID of the parent header or bullet under which to add the new bullet point."`

	Name     string `json:"name,omitempty" jsonschema:"description=Give the code block a custom name; this will function as the uid as well. If it is not unique within the file it will get a suffix."`
	Language string `json:"lang,omitempty" jsonschema:"description=Then code language used in the code block. Can be left empty to make a generic code block with not special formatting or execution."`
	Content  string `json:"content" jsonschema:"description=Text content of the new code block."`
}

func (a AddCodeBlock) Apply(ctx context.Context) (res CodeBlockApplyResult) {
	var of *orgmcp.OrgFile
	of, res.err = orgmcp.OrgFileFromContext(ctx).OkOr(fmt.Errorf(EMPTY_CONTEXT, "orgfile"))
	if res.err != nil {
		return
	}

	var parent Render
	parent, res.err = of.GetUid(a.Parent).OkOr(fmt.Errorf(ITEM_NOT_FOUND, a.Parent))

	var name option.Option[string]
	if strings.TrimSpace(a.Name) == "" {
		name = option.None[string]()
	} else {
		name = option.Some(a.Name)
	}

	var lang option.Option[string]
	if strings.TrimSpace(a.Language) == "" {
		name = option.None[string]()
	} else {
		name = option.Some(a.Language)
	}

	code_block := codeblock.NewCodeBlock(a.Content, name, lang)

	parent.AddChildren(&code_block)

	res.affected = append(res.affected, parent, &code_block)

	return
}

type UpdateCodeBlock struct {
	Method string `json:"method" jsonschema:"enum=update,required=true"`
	Uid    Uid    `json:"parent" jsonschema:"description=UID of the parent header or bullet under which to add the new bullet point."`

	Name     string `json:"name,omitempty" jsonschema:"description=Give the code block a custom name; this will function as the uid as well. If it is not unique within the file it will get a suffix."`
	Language string `json:"lang,omitempty" jsonschema:"description=Then code language used in the code block. Can be left empty to make a generic code block with not special formatting or execution."`
	Content  string `json:"content" jsonschema:"description=Text content of the new code block."`
}

func (u UpdateCodeBlock) Apply(ctx context.Context) (res CodeBlockApplyResult) {
	panic("NOT IMPLEMENTED YET")
}

type RemoveCodeBlock struct {
	Method string `json:"method" jsonschema:"enum=update,required=true"`
	Uid    Uid    `json:"parent" jsonschema:"description=UID of the parent header or bullet under which to add the new bullet point."`
}

func (r RemoveCodeBlock) Apply(ctx context.Context) (res CodeBlockApplyResult) {
	panic("NOT IMPLEMENTED YET")
}

var ManageCodeBlockTool = mcp.GenericTool[ManageCodeBlockInput]{
	Name:        "manage_table",
	Description: "Add, update or remove rows in a table within an Org file.",
	Callback: func(ctx context.Context, input ManageCodeBlockInput, options mcp.FuncOptions) (resp []any, err error) {
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

		affectedMap := map[Uid]Render{}

		for _, item := range input.Items {
			res := item.GetValue().Apply(ctx)

			if res.err != nil {
				resp = append(resp, err)
				continue
			}

			maps.Insert(affectedMap, slice.KeyValueIter(
				res.affected,
				func(t Render) Uid { return t.Uid() },
				func(t Render) Render { return t },
			))
		}

		if !input.HideAffected {
			resp = append(resp, orgmcp.PrintCsv(slices.Collect(maps.Values(affectedMap)), input.Columns))
		}

		return
	},
}
