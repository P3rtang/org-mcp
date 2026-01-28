package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/option"
)

type HeaderInput struct {
	Headers  []HeaderValue `json:"headers"`
	Path     string        `json:"path"`
	ShowDiff bool          `json:"show_diff,omitempty"`
}

type HeaderValue struct {
	Uid     string   `json:"uid"`
	Method  string   `json:"method"`
	Status  string   `json:"status"`
	Content string   `json:"content,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Depth   *int     `json:"depth,omitempty"`
}

var HeaderTool = mcp.Tool{
	Name: "manage_header",
	Description: "Add, remove or update headers in an Org file.\n" +
		"The method parameter defines the action to take: 'add', 'remove', 'update'.\n" +
		"- 'get': Retrieves the header identified by its uid. All other parameters are ignored. Depth however is unique for the get method and returns children up to that depth.\n" +
		"- 'add': Adds a new header at the specified index under the given parent_uid (pass this in the uid field of the function). Requires 'content' parameter.\n" +
		"- 'remove': Removes the header identified by its uid.\n" +
		"- 'update': Updates the header's content, status, or tags. Requires 'content', 'status', or 'tags' parameters.\n\n" +
		"It is recommended to pass uid's as string to the function. While they will almost certainly be numbers, this is not guaranteed.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"headers": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":     "object",
					"required": []string{"method", "uid"},
					"properties": map[string]any{
						"uid": map[string]any{
							"type":        "string",
							"description": "UID of the header to modify, or the parent_uid when adding.",
						},
						"method": map[string]any{
							"type":        "string",
							"enum":        []string{"get", "add", "remove", "update"},
							"description": "The method by which to manage the header.",
						},
						"status": map[string]any{
							"type":        "string",
							"description": "The new status of the header (e.g., TODO, DONE). Required for 'update' and 'add' method. Use 'NONE' to clear status. An empty string will leave the status unchanged during 'update'.",
							"enum":        []string{"TODO", "NEXT", "PROG", "REVW", "DONE", "DELG", "NONE"},
						},
						"content": map[string]any{
							"type":        "string",
							"description": "The content of the header. Required for 'add' and optional for 'update' method.",
						},
						"tags": map[string]any{
							"type":        "array",
							"description": "List of tags to set for the header. Optional for 'update' and 'add' method. Both an empty list and omitting this field will leave tags unchanged during 'update'.",
							"items": map[string]any{
								"type": "string",
							},
						},
						"depth": map[string]any{
							"type":        "integer",
							"description": "The depth to return children headers when using 'get' method. 0 means no children, 1 means direct children only, and so on. If omitted, defaults to 1.",
						},
					},
				},
			},
			"path": map[string]any{
				"type":        "string",
				"description": "The file path to the Org file to modify. It will target the ./.tasks.org by default and you don't have to pass this in unless you want to target a different file.",
				"default":     "./.tasks.org",
			},
			"show_diff": map[string]any{
				"type":        "boolean",
				"description": "Whether to return the diff of changes made to the file. Can be used to inform the user of what changed.",
			},
		},
	},
	Callback: func(args map[string]any, options mcp.FuncOptions) (resp []any, err error) {
		var input HeaderInput

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

		builder := strings.Builder{}

		for _, headerOp := range input.Headers {
			switch headerOp.Method {
			case "get":
				header, ok := orgFile.GetUid(orgmcp.NewUid(headerOp.Uid)).Split()
				if !ok {
					err = errors.New("invalid header UID for get")
					return
				}

				var depth int
				if headerOp.Depth == nil {
					depth = 1
				} else {
					depth = *headerOp.Depth
				}

				header.Render(&builder, depth)
				resp = append(resp, builder.String())
				builder.Reset()
			case "add":
				parent, ok := orgFile.GetUid(orgmcp.NewUid(headerOp.Uid)).Split()
				if !ok {
					err = errors.New("invalid parent UID for adding header")
					return
				}

				var tags option.Option[orgmcp.TagList]
				if len(headerOp.Tags) == 0 {
					tags = option.None[orgmcp.TagList]()
				} else {
					tags = option.Some(orgmcp.TagList(headerOp.Tags))
				}

				newHeader := orgmcp.Header{
					Status:   orgmcp.HeaderStatus(headerOp.Status),
					Progress: option.None[orgmcp.Progress](),
					Tags:     tags,
					Content:  headerOp.Content,
				}
				newHeader.SetLevel(parent.Level() + 1)

				parent.AddChildren(&newHeader)
			case "remove":
				header, ok := orgFile.GetUid(orgmcp.NewUid(headerOp.Uid)).Split()
				if !ok {
					err = errors.New("invalid header UID for removal")
					return
				}

				parent, ok := orgFile.GetUid(header.ParentUid()).Split()
				if !ok {
					err = errors.New("Missing or invalid parent for header removal")
					return
				}

				err = parent.RemoveChildren(orgmcp.NewUid(headerOp.Uid))

				if err != nil {
					return
				}
			case "update":
				header, ok := option.Cast[orgmcp.Render, *orgmcp.Header](orgFile.GetUid(orgmcp.NewUid(headerOp.Uid))).Split()
				if !ok {
					err = errors.New("invalid header UID for update or not a header")
					return
				}

				if headerOp.Content != "" {
					header.Content = headerOp.Content
				}

				if headerOp.Status != "" {
					header.Status = orgmcp.StatusFromString(headerOp.Status)
				}

				if len(headerOp.Tags) != 0 {
					tags := option.Some(orgmcp.TagList(headerOp.Tags))
					header.Tags = tags
				}
			default:
				err = errors.New("invalid method for header management")
			}

			if err != nil {
				return
			}
		}

		diff, err := writeOrgFileToDisk(orgFile, path)
		fmt.Fprintf(os.Stderr, "%t", input.ShowDiff)
		if input.ShowDiff {
			resp = append(resp, diff)
		}

		return
	},
}
