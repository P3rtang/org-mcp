package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/itertools"
)

type TextInputSchema struct {
	Texts        []mcp.OneOf[*TextInputUnion] `json:"texts" jsonschema:"description=The list of text modifications to perform"`
	Path         string                       `json:"path,omitempty" jsonschema:"description=The path to the Org file to modify; if not provided it will default to the current workspace file,required=false"`
	ShowDiff     bool                         `json:"show_diff,omitempty" jsonschema:"description=Whether to show a diff of the changes made; default is false,default=false"`
	ShowAffected *bool                        `json:"show_affected,omitempty" jsonschema:"description=Whether to include the affected items in the response. This will include all items that were modified as well as their children.,default=true,required=false"`
	Columns      []*orgmcp.Column             `json:"columns,omitempty" jsonschema:"description=List of columns to include in the output. If not specified defaults to [UID ; PREVIEW]."`
}

type TextInputUnion struct {
	tag string

	Add    TextInputAdd
	Update TextInputUpdate
	Remove TextInputRemove
}

func NewTextInputUnion[T TextInputAdd | TextInputUpdate | TextInputRemove](input T) *TextInputUnion {
	switch any(input).(type) {
	case TextInputAdd:
		return &TextInputUnion{
			tag: "add",
			Add: any(input).(TextInputAdd),
		}
	case TextInputUpdate:
		return &TextInputUnion{
			tag:    "update",
			Update: any(input).(TextInputUpdate),
		}
	case TextInputRemove:
		return &TextInputUnion{
			tag:    "remove",
			Remove: any(input).(TextInputRemove),
		}
	default:
		panic(fmt.Sprintf("unsupported type for TextInputUnion: %s", reflect.TypeOf(input)))
	}
}

func NewTextInputUnionUpdate(update TextInputUpdate) TextInputUnion {
	return TextInputUnion{
		tag:    "update",
		Update: update,
	}
}

func NewTextInputUnionRemove(remove TextInputRemove) TextInputUnion {
	return TextInputUnion{
		tag:    "remove",
		Remove: remove,
	}
}

func (t *TextInputUnion) Value() any {
	switch t.tag {
	case "add":
		return t.Add
	case "update":
		return t.Update
	default:
		return nil
	}
}

func (t *TextInputUnion) Tag() string {
	return t.tag
}

func (t *TextInputUnion) FromJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch raw["method"] {
	case "add":
		fallthrough
	case nil:
		t.tag = "add"
		return json.Unmarshal(data, &t.Add)
	case "update":
		t.tag = "update"
		return json.Unmarshal(data, &t.Update)
	case "remove":
		t.tag = "remove"
		return json.Unmarshal(data, &t.Remove)
	default:
		return fmt.Errorf("invalid method: %s", raw["method"])
	}
}

type TextInputAdd struct {
	Method  string `json:"method" jsonschema:"description=Add new text content under the specified parent element.,enum=add"`
	Parent  string `json:"parent" jsonschema:"description=The UID of the parent element under which the text will be added. This can be either a header or a bullet point."`
	Content string `json:"content" jsonschema:"description=The text content to add. Newlines will result in multiple plain text elements being added, one for each line of text."`
}

func (t *TextInputAdd) Apply(ctx context.Context, of *orgmcp.OrgFile) (res ApplyResult) {
	res.affectedItems = make(map[orgmcp.Uid]orgmcp.Render)
	parentUid, ok := of.GetUid(orgmcp.NewUid(t.Parent)).Split()

	if !ok {
		res.err = fmt.Errorf("Parent uid %s not found.", t.Parent)
		return
	}

	for c := range strings.SplitSeq(t.Content, "\n") {
		newPlainText := orgmcp.NewPlainText(c)
		parentUid.AddChildren(&newPlainText)

		res.affectedItems[newPlainText.Uid()] = &newPlainText
	}

	return
}

type TextInputUpdate struct {
	Uid     string `json:"uid" jsonschema:"description=The UID of the element to modify or remove."`
	Method  string `json:"method" jsonschema:"description=Update the content of a text element.,enum=update"`
	Content string `json:"content,omitempty" jsonschema:"description=The new content of the text element."`
}

func (t *TextInputUpdate) Apply(ctx context.Context, of *orgmcp.OrgFile) (res ApplyResult) {
	res.affectedItems = make(map[orgmcp.Uid]orgmcp.Render)
	selected, ok := of.GetUid(orgmcp.NewUid(t.Uid)).Split()

	if !ok {
		res.err = fmt.Errorf("Item with uid %s not found in %s.", t.Uid, of.Name())
		return
	}
	if strings.Contains(t.Content, "\n") {
		res.err = fmt.Errorf("Content with newlines is not allowed for the update method, these will be replaced with spaces. If you want to add content with newlines you should use the 'add' method to add new text elements for each line of text.")
		t.Content = strings.ReplaceAll(t.Content, "\n", " ")
	}

	if plain, ok := selected.(*orgmcp.PlainText); ok {
		plain.SetContent(t.Content)
		res.affectedItems[plain.Uid()] = plain
	} else {
		res.err = fmt.Errorf("Uid %s is not a plain text element, cannot update content", t.Uid)
	}

	return
}

type TextInputRemove struct {
	Uid    string `json:"uid" jsonschema:"description=The UID of the element to modify or remove."`
	Method string `json:"method" jsonschema:"description=Remove the text element.,enum=remove"`
}

func (t *TextInputRemove) Apply(ctx context.Context, of *orgmcp.OrgFile) (res ApplyResult) {
	res.affectedItems = make(map[orgmcp.Uid]orgmcp.Render)
	selected, ok := of.GetUid(orgmcp.NewUid(t.Uid)).Split()

	if !ok {
		res.err = fmt.Errorf("Item with uid %s not found in %s.", t.Uid, of.Name())
		return
	}

	p_uid := selected.ParentUid()
	if parent, ok := of.GetUid(p_uid).Split(); ok {
		parent.RemoveChildren(selected.Uid())
		res.affectedItems[p_uid] = parent
	} else {
		res.err = fmt.Errorf("Parent with uid %s not found for item with uid %s", p_uid, t.Uid)
	}

	return
}

var TextTool = mcp.GenericTool[TextInputSchema]{
	Name: "manage_text",
	Description: `
Add, update or remove text content in an Org file.
Plain text is a special type of content that cannot contain any nested elements.
It is solely used for storing text content within a header or bullet point.

Most of the time it will be more correct to use a bullet point to store text content as it allows for better organization and structuring of the content.
But there are situations where plain text is more appropriate, such as large block of text without structure, like github issues etc.

## Methods
` +
		"`add`: Adds new text content to the specified parent element. The parent is passed via the uid parameter.\n" +
		"`update`: Updates the text content of the specified element. The element is identified by its uid.\n" +
		"`remove`: Removes the text content of the specified element. The element is identified by its uid.\n" +
		`
## UID Constructions
You can target either a header or a bullet point when adding text content. The uid will be the parent itself.
Otherwise the uid construction is similar to the bullet tool.
The text index is relative to the parent element, you can have multiple text elements under the same parent.
The index is based on newline separation, but is deligneated by the parent element.

## Diff
You always have the options with any modification to show a diff of the changes made.
This can inform both you as well as the user about what exactly a tool call changed, and always you to undo changes if needed.
` +
		"`parent_uid + .t + text_index`\n",
	Callback: func(ctx context.Context, input TextInputSchema, options mcp.FuncOptions) (resp []any, err error) {
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

		for _, mt := range input.Texts {
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
