package tools

import (
	"encoding/json"

	"github.com/p3rtang/org-mcp/mcp"
)

type InputSchema struct {
	Path string `json:"path"`
}

var StatusTool = mcp.Tool{
	Name: "status_overview",
	Description: "Provides an overview of task statuses in the Org file.\n" +
		"It counts the number of tasks in each status category (TODO, NEXT, PROG, REVW, DONE, DELG).\n" +
		"You will also receive all the uid's of the headers for each category in a list.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The file path to the Org file to analyze.",
			},
		},
	},

	Callback: func(args map[string]any, options mcp.FuncOptions) (resp []any, err error) {
		var input InputSchema

		bytes, err := json.Marshal(args)
		if err != nil {
			return
		}

		err = json.Unmarshal(bytes, &input)
		if err != nil {
			return
		}

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
		resp = []any{orgFile.GetStatusOverview()}

		_, err = writeOrgFileToDisk(orgFile, path)

		return
	},
}
