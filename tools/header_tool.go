package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/itertools"
	"github.com/p3rtang/org-mcp/utils/option"
)

type HeaderInput struct {
	Headers      []mcp.OneOf[*HeaderInputUnion] `json:"headers" jsonschema:"description=List of header operations to perform. Multiple operations can be performed in a single call."`
	Path         string                         `json:"path,omitempty" jsonschema:"description=The file path to the Org file to modify. It will target the ./.tasks.org by default and you don't have to pass this in unless you want to target a different file.,required=false"`
	ShowDiff     bool                           `json:"show_diff,omitempty" jsonschema:"description=Whether to return the diff of changes made to the file. Can be used to inform the user of what changed.,required=false"`
	ShowAffected *bool                          `json:"show_affected,omitempty" jsonschema:"description=Whether to include the affected items in the response. This will include all items that were modified as well as their children.,default=true,required=false"`
	Columns      []*orgmcp.Column               `json:"columns,omitempty" jsonschema:"description=List of columns to include in the output. If not specified defaults to [UID | PREVIEW]."`
}

type HeaderInputUnion struct {
	tag string

	Add    HeaderInputAdd
	Update HeaderInputUpdate
	Remove HeaderInputRemove
}

func NewHeaderInputUnion[T HeaderInputAdd | HeaderInputUpdate | HeaderInputRemove](input T) *HeaderInputUnion {
	switch any(input).(type) {
	case HeaderInputAdd:
		return &HeaderInputUnion{
			tag: "add",
			Add: any(input).(HeaderInputAdd),
		}
	case HeaderInputUpdate:
		return &HeaderInputUnion{
			tag:    "update",
			Update: any(input).(HeaderInputUpdate),
		}
	case HeaderInputRemove:
		return &HeaderInputUnion{
			tag:    "remove",
			Remove: any(input).(HeaderInputRemove),
		}
	default:
		panic(fmt.Sprintf("unsupported type for HeaderInputUnion: %s", fmt.Sprintf("%T", input)))
	}
}

func (u HeaderInputUnion) Tag() string {
	return u.tag
}

func (u HeaderInputUnion) Value() any {
	switch u.tag {
	case "add":
		return u.Add
	case "update":
		return u.Update
	case "remove":
		return u.Remove
	default:
		return nil
	}
}

func (u *HeaderInputUnion) FromJSON(data []byte) error {
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
	default:
		return errors.New("invalid method for header operation")
	}
}

type HeaderInputAdd struct {
	Method  string   `json:"method" jsonschema:"description=Add a new header.,enum=add"`
	Parent  string   `json:"parent" jsonschema:"description=UID of the parent header under which to add the new header."`
	Content string   `json:"content" jsonschema:"description=The content of the new header."`
	Status  string   `json:"status,omitempty" jsonschema:"description=The status of the new header (e.g. TODO; DONE). Use 'NONE' or omit the field to leave status empty.,enum=TODO;NEXT;PROG;REVW;DONE;DELG;NONE"`
	Tags    []string `json:"tags,omitempty" jsonschema:"description=List of tags to set for the new header. An empty list or omitting this field will leave tags empty."`
}

func (h HeaderInputAdd) Apply(ctx context.Context, of *orgmcp.OrgFile) (res ApplyResult) {
	res.affectedItems = make(map[orgmcp.Uid]orgmcp.Render)

	if h.Content == "" {
		res.err = errors.New("Content cannot be empty when adding a header.")
		return
	}

	parent, ok := of.GetUid(orgmcp.NewUid(h.Parent)).Split()
	if !ok {
		res.err = fmt.Errorf("Parent UID %s not found.", h.Parent)
		return
	}

	header := orgmcp.NewHeader(
		orgmcp.HeaderStatus(h.Status),
		h.Content,
	)

	if len(h.Tags) != 0 {
		header.Tags = option.Some(orgmcp.TagList(h.Tags))
	}

	parent.AddChildren(&header)

	res.affectedItems[parent.Uid()] = &header

	return
}

type HeaderInputUpdate struct {
	Method  string   `json:"method" jsonschema:"description=Update an existing header.,enum=update"`
	Uid     string   `json:"uid" jsonschema:"description=UID of the header to update."`
	Content string   `json:"content,omitempty" jsonschema:"description=The new content of the header. Omit this field to keep the content unchanged."`
	Status  string   `json:"status,omitempty" jsonschema:"description=The new status of the header (e.g. TODO; DONE). Use 'NONE' to clear status. An empty string or omitting this field will leave status unchanged.,enum=TODO;NEXT;PROG;REVW;DONE;DELG;NONE"`
	Tags    []string `json:"tags,omitempty" jsonschema:"description=List of tags to set for the header. Both an empty list and omitting this field will leave tags unchanged."`
}

func (h HeaderInputUpdate) Apply(ctx context.Context, of *orgmcp.OrgFile) (res ApplyResult) {
	res.affectedItems = make(map[orgmcp.Uid]orgmcp.Render)

	header, ok := option.Cast[orgmcp.Render, *orgmcp.Header](of.GetUid(orgmcp.NewUid(h.Uid))).Split()
	if !ok {
		res.err = fmt.Errorf("Header with UID %s not found.", h.Uid)
		return
	}

	if h.Content != "" {
		header.SetContent(h.Content)
	}

	if h.Status != "" {
		header.SetStatus(orgmcp.HeaderStatus(h.Status))
	}

	if len(h.Tags) != 0 {
		header.Tags = option.Some(orgmcp.TagList(h.Tags))
	}

	res.affectedItems[header.Uid()] = header

	return
}

type HeaderInputRemove struct {
	Method string `json:"method" jsonschema:"description=Remove an existing header.,enum=remove"`
	Uid    string `json:"uid" jsonschema:"description=UID of the header to remove."`
}

func (h HeaderInputRemove) Apply(ctx context.Context, of *orgmcp.OrgFile) (res ApplyResult) {
	res.affectedItems = make(map[orgmcp.Uid]orgmcp.Render)

	header, ok := of.GetUid(orgmcp.NewUid(h.Uid)).Split()
	if !ok {
		res.err = fmt.Errorf("Header with UID %s not found.", h.Uid)
		return
	}

	parent, ok := of.GetUid(header.ParentUid()).Split()
	if !ok {
		res.err = fmt.Errorf("Parent of header with UID %s not found.", h.Uid)
		return
	}

	err := parent.RemoveChildren(header.Uid())
	if err != nil {
		res.err = err
		return
	}

	res.affectedItems[parent.Uid()] = parent

	return
}

var HeaderTool = mcp.GenericTool[HeaderInput]{
	Name: "manage_header",
	Description: "Add; remove or update headers in an Org file.\n" +
		"The method parameter defines the action to take: 'add'; 'remove'; 'update'.\n" +
		"For any method you can use a depth parameter to specify how many levels of children to return.\n" +
		"- 'add': Adds a new header at the specified index under the given parent_uid (pass this in the uid field of the function). Requires 'content' parameter.\n" +
		"- 'remove': Removes the header identified by its uid.\n" +
		"- 'update': Updates the header's content; status; or tags. Requires 'content'; 'status'; or 'tags' parameters.\n\n" +
		"It is recommended to pass uid's as string to the function. While they will almost certainly be numbers; this is not guaranteed.",
	Callback: func(ctx context.Context, input HeaderInput, options mcp.FuncOptions) (resp []any, err error) {
		var path string
		if input.Path == "" {
			path = options.DefaultPath
		} else {
			path = input.Path
		}

		orgFile, err := mcp.LoadOrgFile(ctx, path)
		if err != nil {
			return
		}

		affectedCount := 0
		affectedItems := map[orgmcp.Uid]orgmcp.Render{}

		for _, mt := range input.Headers {
			var res ApplyResult

			switch mt.Value.Tag() {
			case "add":
				res = mt.Value.Add.Apply(ctx, &orgFile)
			case "update":
				res = mt.Value.Update.Apply(ctx, &orgFile)
			case "remove":
				res = mt.Value.Remove.Apply(ctx, &orgFile)
			}

			if res.err != nil {
				resp = append(resp, res.err.Error())
			}

			maps.Copy(affectedItems, res.affectedItems)
			affectedCount += len(res.affectedItems)
		}

		ordered := []orgmcp.Render{}

		if input.ShowAffected == nil || *input.ShowAffected == true {
			locationTable := orgFile.BuildLocationTable()
			ordered = append(ordered, itertools.Collect(maps.Values(affectedItems))...)

			slices.SortFunc(ordered, func(a, b orgmcp.Render) int {
				return (*locationTable)[a.Uid()] - (*locationTable)[b.Uid()]
			})

			resp = append(resp, orgmcp.PrintCsv(ordered, input.Columns))
			resp = append(resp, map[string]any{
				"affected_count": affectedCount,
			})
		}

		diff, err := mcp.WriteOrgFileToDisk(ctx, orgFile, path)
		if input.ShowDiff {
			resp = append(resp, diff)
		}

		return
	},
}
