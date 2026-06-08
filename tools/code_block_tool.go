package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"reflect"
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

	Add    AddCodeBlock
	Update UpdateCodeBlock
	Remove RemoveCodeBlock
}

func NewManageCodeBlockInputUnion[T AddCodeBlock | UpdateCodeBlock | RemoveCodeBlock](t T) *ManageCodeBlockInputUnion {
	switch any(t).(type) {
	case AddCodeBlock:
		return &ManageCodeBlockInputUnion{tag: "add", Add: any(t).(AddCodeBlock)}
	case UpdateCodeBlock:
		return &ManageCodeBlockInputUnion{tag: "update", Update: any(t).(UpdateCodeBlock)}
	case RemoveCodeBlock:
		return &ManageCodeBlockInputUnion{tag: "remove", Remove: any(t).(RemoveCodeBlock)}
	default:
		panic("unreachable")
	}
}

func (mcb *ManageCodeBlockInputUnion) Tag() string { return mcb.tag }

func (mcb *ManageCodeBlockInputUnion) Value() CodeBlockApplicableTool {
	switch mcb.Tag() {
	case "add":
		return mcb.Add
	case "update":
		return mcb.Update
	case "remove":
		return mcb.Remove
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
		return json.Unmarshal(data, &mcb.Add)
	case "update":
		mcb.tag = "update"
		return json.Unmarshal(data, &mcb.Update)
	case "remove":
		mcb.tag = "remove"
		return json.Unmarshal(data, &mcb.Remove)
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
	Parent Uid    `json:"parent" jsonschema:"description=UID of the parent header or bullet under which to add the new code block."`

	Name     string `json:"name,omitempty" jsonschema:"description=Give the code block a custom name; Serves as both the name and the uid if specified, otherwise a numeric index based uid will be used for future method calls. If it is not unique within the file it will get a suffix."`
	Language string `json:"lang,omitempty" jsonschema:"description=The code language used in the code block. Can be left empty to make a generic code block with not special formatting or execution."`
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
		lang = option.None[string]()
	} else {
		lang = option.Some(a.Language)
	}

	code_block := codeblock.NewCodeBlock(a.Content, name, lang)

	parent.AddChildren(&code_block)

	res.affected = append(res.affected, parent, &code_block)

	return
}

type UpdateCodeBlock struct {
	Method string `json:"method" jsonschema:"description=All fields in the update method are optional and only fields explicitly set will be updated.,enum=update,required=true"`
	Uid    Uid    `json:"uid" jsonschema:"description=UID of the CodeBlock to edit."`

	Name     string `json:"name,omitempty" jsonschema:"description=Update the name of the code block; when setting the name it will also change the uid of this code block. If it is not unique within the file it will get a suffix."`
	Language string `json:"lang,omitempty" jsonschema:"description=The code language used in the code block. Can be left empty to make a generic code block with not special formatting or execution."`
	Content  string `json:"content,omitempty" jsonschema:"description=Text content of the code block."`
}

func (u UpdateCodeBlock) Apply(ctx context.Context) (res CodeBlockApplyResult) {
	var of *orgmcp.OrgFile
	of, res.err = orgmcp.OrgFileFromContext(ctx).OkOr(fmt.Errorf(EMPTY_CONTEXT, "orgfile"))
	if res.err != nil {
		return
	}

	var item Render
	item, res.err = of.GetUid(u.Uid).OkOr(fmt.Errorf(ITEM_NOT_FOUND, u.Uid))

	if res.err != nil {
		return
	}

	if code_block, ok := item.(*codeblock.CodeBlock); ok {
		if u.Content != "" {
			code_block.SetContent(u.Content)
		}

		if u.Language != "" {
			code_block.SetLanguage(u.Language)
		}

		if u.Name != "" {
			code_block.SetName(u.Name)
		}

		res.affected = append(res.affected, code_block)
	} else {
		res.err = fmt.Errorf(WRONG_TYPE, u.Uid, "CodeBlock", reflect.TypeOf(item))
	}

	return
}

type RemoveCodeBlock struct {
	Method string `json:"method" jsonschema:"enum=remove,required=true"`
	Uid    Uid    `json:"uid" jsonschema:"description=UID of the CodeBlock to remove."`
}

func (r RemoveCodeBlock) Apply(ctx context.Context) (res CodeBlockApplyResult) {
	var of *orgmcp.OrgFile
	of, res.err = orgmcp.OrgFileFromContext(ctx).OkOr(fmt.Errorf(EMPTY_CONTEXT, "orgfile"))
	if res.err != nil {
		return
	}

	var item Render
	item, res.err = of.GetUid(r.Uid).OkOr(fmt.Errorf(ITEM_NOT_FOUND, r.Uid))

	if res.err != nil {
		return
	}

	var parent Render
	parent, res.err = of.GetUid(item.ParentUid()).OkOr(fmt.Errorf(PARENT_NOT_FOUND, r.Uid))
	if res.err != nil {
		return
	}

	res.err = parent.RemoveChildren(r.Uid)

	return
}

var ManageCodeBlockTool = mcp.GenericTool[ManageCodeBlockInput]{
	Name: "manage_codeblock",
	Description: `
Add, update, remove, or execute codeblock items in an Org file.

## Code Block Structure
Code blocks are delimited by '#+BEGIN_SRC' and '#+END_SRC'. They support:
- '#+NAME:' for stable identification
- Language tag after '#+BEGIN_SRC'
- Arbitrary content between delimiters

## Methods

### add
Adds a new codeblock under the specified parent header or bullet.
- 'parent' (required): UID of the parent
- 'content' (required): The code block content
- 'name' (optional): Sets #+NAME for stable UID
- 'lang' (optional): Language tag (python, javascript, bash, etc.)

### update
Updates an existing codeblock. All fields optional - only provided fields are updated.
- 'uid' (required): UID of the codeblock to update
- 'content' (optional): New content
- 'name' (optional): Rename (changes UID)
- 'lang' (optional): Change language

### remove
Removes a codeblock from the file.
- 'uid' (required): UID of the codeblock to remove

### execute (WIP: not implemented yet)
Executes the codeblock content and returns results.
- 'uid' (required): UID of the codeblock to execute
- 'async' (optional, default false): If true, returns taskId immediately for polling
- 'timeout' (optional, default 60): Max execution time in seconds
- 'insert_results' (optional, default true): Whether to insert #+RESULTS block below source

**Sync mode** (async=false): Blocks until execution complete, returns stdout/stderr/exit code directly.

**Async mode** (async=true): Returns CreateTaskResult with taskId. Poll tasks/get for completion. Use for long-running code (e.g., Playwright tests with browser setup).

**Supported languages**: python, javascript, bash

**Execution environment**: Docker containers with resource limits. 60s max timeout.

**Results format**:
- stdout: Standard output
- stderr: Standard error (separate from stdout)
- exitCode: Process exit code
- executionTimeMs: How long execution took
`,
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
			return resp, err
		}

		affectedMap := map[Uid]Render{}

		for _, item := range input.Items {
			res := item.GetValue().Apply(ctx)

			if res.err != nil {
				resp = append(resp, res.err)
				continue
			}

			maps.Insert(affectedMap, slice.KeyValueIter(
				res.affected,
				func(t Render) Uid { return t.Uid() },
				func(t Render) Render { return t },
			))
		}

		if !input.HideAffected && len(affectedMap) > 0 {
			if len(input.Columns) == 0 {
				input.Columns = append(input.Columns, &orgmcp.ColUidValue, &orgmcp.ColPreviewValue)
			}

			fmt.Fprintf(os.Stderr, "%#v", affectedMap)
			resp = append(resp, orgmcp.PrintCsv(slices.Collect(maps.Values(affectedMap)), input.Columns))
		}

		var diff string
		diff, err = mcp.WriteOrgFileToDisk(ctx, orgFile, path)

		if input.ShowDiff {
			resp = append(resp, diff)
		}

		return
	},
}
