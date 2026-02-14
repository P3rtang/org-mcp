package tools

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/itertools"
)

type TextInputSchema struct {
	Texts        []TextInputValue `json:"texts" jsonschema:"description=The list of text modifications to perform"`
	Path         string           `json:"path,omitempty" jsonschema:"description=The path to the Org file to modify; if not provided it will default to the current workspace file,required=false"`
	ShowDiff     bool             `json:"show_diff,omitempty" jsonschema:"description=Whether to show a diff of the changes made; default is false,required=false"`
	ShowAffected *bool            `json:"show_affected,omitempty" jsonschema:"description=Whether to include the affected items in the response. This will include all items that were modified as well as their children.,default=true,required=false"`
	Columns      []*orgmcp.Column `json:"columns,omitempty" jsonschema:"description=List of columns to include in the output. If not specified defaults to [UID ; PREVIEW]."`
}

type TextInputValue struct {
	Uid     string `json:"uid" jsonschema:"description=The UID of the element to modify. This can be either a header or a bullet point. When adding text content the UID will refer to the parent element under which the text will be added."`
	Method  string `json:"method" jsonschema:"description=The method of modification to perform (add; update or remove). When adding text content the method must be 'add'.;enum=add;update;remove"`
	Content string `json:"content,omitempty" jsonschema:"description=The text content to add or update. When using the 'remove' method this field is ignored."`
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
	Callback: func(input TextInputSchema, options mcp.FuncOptions) (resp []any, err error) {
		var path string
		if input.Path == "" {
			path = options.DefaultPath
		} else {
			path = input.Path
		}

		orgFile, err := mcp.LoadOrgFile(path)
		if err != nil {
			return
		}

		affectedCount := 0
		affectedItems := map[orgmcp.Uid]orgmcp.Render{}

		for _, mt := range input.Texts {
			var selected orgmcp.Render
			var ok bool
			if selected, ok = orgFile.GetUid(orgmcp.NewUid(mt.Uid)).Split(); !ok {
				resp = append(resp, fmt.Sprintf("Uid %s not found in %s", mt.Uid, path))
				continue
			}

			switch mt.Method {
			case "add":
				if strings.Contains(mt.Content, "\n") {
					resp = append(resp, "Content for update method should not contain newlines. You should update each text element separately. As a fallback the newlines will be replaced with spaces.")
					mt.Content = strings.ReplaceAll(mt.Content, "\n", " ")
				}

				newPlainText := orgmcp.NewPlainText(mt.Content)
				selected.AddChildren(&newPlainText)

				affectedItems[selected.Uid()] = selected
				for _, child := range selected.ChildrenRec(-1) {
					affectedItems[child.Uid()] = child
				}

				affectedCount += 1
			case "update":
				if strings.Contains(mt.Content, "\n") {
					resp = append(resp, "Content for update method should not contain newlines. You should update each text element separately. As a fallback the newlines will be replaced with spaces.")
					mt.Content = strings.ReplaceAll(mt.Content, "\n", " ")
				}

				if plain, ok := selected.(*orgmcp.PlainText); ok {
					plain.SetContent(mt.Content)
					affectedItems[plain.Uid()] = plain
					affectedCount += 1
				}
			case "remove":
				p_uid := selected.ParentUid()
				if parent, ok := orgFile.GetUid(p_uid).Split(); ok {
					parent.RemoveChildren(selected.Uid())
				}

				affectedItems[selected.Uid()] = selected
				for _, child := range selected.ChildrenRec(-1) {
					affectedItems[child.Uid()] = child
				}

				affectedCount += 1
			}
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

		diff, err := writeOrgFileToDisk(orgFile, path)
		if input.ShowDiff {
			resp = append(resp, diff)
		}

		return
	},
}
