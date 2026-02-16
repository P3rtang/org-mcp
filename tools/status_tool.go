package tools

import (
	"context"

	"github.com/p3rtang/org-mcp/mcp"
)

type StatusInputSchema struct {
	Path string `json:"path,omitempty" jsonschema:"description=The file path to the Org file to modify. It will target the ./.tasks.org by default and you don't have to pass this in unless you want to target a different file.,required=false"`
}

var StatusTool = mcp.GenericTool[StatusInputSchema]{
	Name: "status_overview",
	Description: `
Provides an overview of task statuses in the Org file.
It counts the number of tasks in each status category (TODO, NEXT, PROG, REVW, DONE, DELG).
You will also receive all the uid's of the headers for each category in a list.

This tool will also return an overview of all tags used in the Org file and the count of how many times each tag is used.
Remember however that tags are recursive children will inherit the tags of their parents,
so a header with the tag "project" and a child header without any tags will still be counted as having the "project" tag.
`,

	Callback: func(ctx context.Context, input StatusInputSchema, options mcp.FuncOptions) (resp []any, err error) {
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

		resp = []any{map[string]any{
			"status_overview": orgFile.GetStatusOverview(),
			"tag_overview":    orgFile.GetTagOverview(),
		}}

		_, err = mcp.WriteOrgFileToDisk(ctx, orgFile, path)

		return
	},
}
