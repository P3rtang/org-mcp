package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/tools"
	"github.com/p3rtang/org-mcp/utils/slice"
)

// loadOrgFile loads an OrgFile from the given file path.
// It opens the file, reads it using OrgFileFromReader, and returns the result.
func loadOrgFile(filePath string) (orgmcp.OrgFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return orgmcp.OrgFile{}, err
	}
	defer file.Close()

	return orgmcp.OrgFileFromReader(file).Split()
}

// writeOrgFileToDisk renders the OrgFile and writes it to the provided file path.
func writeOrgFileToDisk(of orgmcp.OrgFile, filePath string) (err error) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	builder := strings.Builder{}
	of.Render(&builder, -1)

	_, err = file.WriteString(builder.String())

	return
}

func main() {
	defaultPath := ".tasks.org"

	// Setup logging to stderr so it doesn't interfere with stdout (which is used for MCP protocol)
	logger := log.New(os.Stderr, "[org-mcp] ", log.LstdFlags|log.Lshortfile)

	// Create a message sender that encodes and sends JSON to stdout
	// Using a persistent encoder for better performance and proper flushing
	encoder := json.NewEncoder(os.Stdout)
	sender := mcp.NewMessageSender(func(msg any) error {
		return encoder.Encode(msg)
	})

	// Create and run the MCP server
	server := mcp.NewServer(os.Stdin, sender, logger)

	server.AddTool(
		mcp.Tool{
			Name: "get_header_by_status",
			Description: "Get all headers that have the input status set.\n" +
				"If a subheader is specified, only its descendant headers are considered." +
				"To view the contents of the returned header use the view_header tool.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"status": map[string]any{
						"type":        "string",
						"description": "The status the headers should have.",
						"enum":        []string{"TODO", "NEXT", "PROG", "DONE"},
					},
					"subheader": map[string]any{
						"type":        "number",
						"description": "The subheader to filter by.",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "An optional file path, will default to the configured org file.",
					},
				},
			},
		},
		func(args map[string]any, _ mcp.FuncOptions) (resp any, err error) {
			response := map[string]any{}

			filePath, ok := args["path"].(string)

			if !ok || filePath == "" {
				filePath = defaultPath
			}

			of, err := loadOrgFile(filePath)
			if err != nil {
				return
			}

			status := orgmcp.StatusFromString(args["status"].(string))

			if status == orgmcp.None {
				return nil, errors.New("invalid status")
			}

			response["headers"] = slice.Map(of.GetHeaderByStatus(status), func(h *orgmcp.Header) map[string]any {
				builder := strings.Builder{}
				h.Render(&builder, 1)

				return map[string]any{
					"uid":        h.Uid(),
					"content":    builder.String(),
					"parent_uid": h.ParentUid(),
				}
			})

			err = writeOrgFileToDisk(of, filePath)
			if err != nil {
				return
			}

			resp = response

			return
		},
	)

	server.AddTool(
		mcp.Tool{
			Name: "view_header",
			Description: "View any specific headers by their uid.\n" +
				"Use this to render the header itself, its content including properties and schedules.\n" +
				"But also all of its children recursively up to a given depth.\n" +
				"Depth 0 will also include just the header content of the direct children of the header, but hide their content.\n" +
				"You should think of headers as a directory containing other dirs (sub-headers), but also files and metadata.\n\n" +
				"While not recommended it is possible to request the whole tasks file at once. You can achieve this be setting the uid to 0 and a depth of -1",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"uid": map[string]any{
						"type":        "number",
						"description": "The uid of the header to view. The root of the file will always be uid 0.",
					},
					"depth": map[string]any{
						"type":        "number",
						"description": "The depth to which children should be rendered. A depth of -1 will return the whole content as if there is no depth cutoff.",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "An optional file path, will default to the configured org file.",
					},
				},
			},
		}, func(args map[string]any, options mcp.FuncOptions) (resp any, err error) {
			filePath, ok := args["path"].(string)

			if !ok || filePath == "" {
				filePath = options.DefaultPath
			}

			of, err := loadOrgFile(filePath)
			if err != nil {
				return
			}

			fmt.Fprintf(os.Stderr, "\nuid: %v\n", args["uid"])
			// Get the index (uid) from the arguments
			uid := -1
			if u, ok := args["uid"].(float64); ok {
				uid = int(u)
			}

			if uid == -1 {
				return nil, errors.New("invalid or missing uid parameter")
			}

			// Get the depth parameter, default to 1 if not provided
			depth := 1
			if depthVal, ok := args["depth"].(float64); ok {
				depth = int(depthVal)
			}

			// Retrieve the header by uid
			header, ok := of.GetUid(orgmcp.NewUid(uid)).Split()
			if !ok {
				return nil, fmt.Errorf("header with uid %v not found", uid)
			}

			// Render the header with the specified depth
			builder := strings.Builder{}
			header.Render(&builder, depth)

			response := map[string]any{
				"uid":        header.Uid(),
				"content":    builder.String(),
				"parent_uid": header.ParentUid(),
			}

			_ = writeOrgFileToDisk(of, filePath)

			return response, nil
		},
	)

	server.AddTool(
		mcp.Tool{
			Name: "set_header_status",
			Description: "Set/Cycle the status of a header. As described passing a status will set it, otherwise it will cycle as documented in the status field.\n" +
				"You should always try and update the status to reflect reality as much as possible.\n" +
				"Make sure you also have the most recent status at the moment, don't forget this is a collaborative document.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"uid": map[string]any{
						"description": "Pass an array to modify multiple headers, when increasing status it will increase it separately for every item.\n" +
							"It won't synchronise them if they are not synchronised currently, you can safely bump multiple distinct statuses.\n" +
							"Also don't forget to toggle any children or parent headers that should reflect downstream changes",
						"type": "array",
						"items": map[string]any{
							"type": "number",
						},
					},
					"status": map[string]any{
						"type":        "string",
						"description": "The new status of the header, leave empty to cycle through the statusses. The order is NONE -> TODO -> NEXT -> PROG -> DONE -> NONE",
						"enum":        []string{"TODO", "NEXT", "PROG", "DONE"},
					},
					"path": map[string]any{
						"type":        "string",
						"description": "An optional file path, will default to the configured org file.",
					},
				},
			},
		}, func(args map[string]any, _ mcp.FuncOptions) (resp any, err error) {
			response := []any{}
			uids := []int{}

			if uids, ok := args["uid"].([]any); ok {
				for _, uid := range uids {
					if u, ok := uid.(string); ok {
						uids = append(uids, orgmcp.NewUid(u))
					} else {
						err = fmt.Errorf("invalid uid type in array")
						return
					}
				}
			} else {
				err = fmt.Errorf("invalid uid type")
				return
			}

			filePath, ok := args["path"].(string)

			if !ok || filePath == "" {
				filePath = defaultPath
			}

			org_file, err := loadOrgFile(filePath)
			if err != nil {
				return
			}

			for _, uid := range uids {
				var r orgmcp.Render
				var h *orgmcp.Header
				if r, ok = org_file.GetUid(orgmcp.NewUid(uid)).Split(); !ok {
					err = fmt.Errorf("Header #%d not found", uid)
					return
				}

				if h, ok = r.(*orgmcp.Header); !ok {
					err = fmt.Errorf("Item #%d is not a header", uid)
					return
				}

				s, ok := args["status"].(string)
				if !ok {
					h.Status = h.Status.GetNext()
				} else {
					h.Status = orgmcp.StatusFromString(s)
				}
			}

			err = writeOrgFileToDisk(org_file, filePath)

			return response, err
		},
	)

	server.AddTool(
		mcp.Tool{
			Name:        "create_subheader",
			Description: "Create a new subheader under an existing header by parent UID. Allows specifying title, optional content, and status.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []string{"parent_uid", "title"},
				"properties": map[string]any{
					"parent_uid": map[string]any{
						"type":        "number",
						"description": "The UID of the parent header to which the subheader will be added.",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "Title/content of new header. Must not be empty.",
					},
					"status": map[string]any{
						"type":        "string",
						"enum":        []string{"TODO", "NEXT", "PROG", "DONE"},
						"description": "Optional initial status. Defaults to no status.",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "Optional file path to org file. Defaults to configured org file.",
					},
				},
			},
		},
		func(args map[string]any, _ mcp.FuncOptions) (resp any, err error) {
			filePath, ok := args["path"].(string)
			if !ok || filePath == "" {
				filePath = defaultPath
			}

			of, err := loadOrgFile(filePath)
			if err != nil {
				return
			}

			parentUid, ok := args["parent_uid"].(string)
			if !ok {
				return nil, errors.New("Invalid or missing parent_uid parameter")
			}

			title, ok := args["title"].(string)
			if !ok || strings.TrimSpace(title) == "" {
				return nil, errors.New("Title must not be empty.")
			}

			statusStr, _ := args["status"].(string)

			status := orgmcp.StatusFromString(statusStr)
			if status == orgmcp.None {
				status = orgmcp.Todo
			}

			render, ok := of.GetUid(orgmcp.NewUid(parentUid)).Split()
			if !ok {
				return nil, errors.New("Parent header not found.")
			}

			parentHeader, ok := render.(*orgmcp.Header)
			if !ok {
				return nil, errors.New("Parent UID is not a header.")
			}

			newHeader := parentHeader.CreateSubheader(title, status)
			builder := strings.Builder{}
			of.Render(&builder, -1)

			file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			file.WriteString(builder.String())

			builder.Reset()
			parentHeader.Render(&builder, 1)
			resp = map[string]any{
				"uid":        newHeader.Uid(),
				"parent_uid": parentUid,
				"content":    builder.String(),
			}

			return
		},
	)

	server.AddTool(tools.BulletTool, tools.BulletFunc)

	server.AddTool(mcp.Tool{
		Name: "vector_search",
		Description: "Perform a vector search on all headers in the org file based on the provided query string. " +
			"Returns the top N most relevant headers.\n" +
			"It is optimal to include as much information as possible in the query, overflowing the text limit is hard. " +
			"So include context like timeframes, people involved, locations, etc. to get the best results.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The search query string.",
				},
				"top_n": map[string]any{
					"type":        "number",
					"description": "The number of top relevant headers to return.",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "An optional file path, will default to the configured org file.",
				},
			},
		},
	}, func(args map[string]any, _ mcp.FuncOptions) (resp any, err error) {
		filePath, ok := args["path"].(string)
		if !ok || strings.TrimSpace(filePath) == "" {
			filePath = defaultPath
		}

		query, ok := args["query"].(string)
		if !ok || strings.TrimSpace(query) == "" {
			return nil, errors.New("invalid or missing query parameter")
		}

		top_n, ok := args["top_n"].(float64)
		if !ok || top_n <= 0 {
			top_n = 0
		}

		of, err := loadOrgFile(filePath)
		if err != nil {
			return
		}

		headers, err := of.VectorSearch(args["query"].(string), int(top_n))

		resp = map[string]any{
			"results": slice.Map(headers, func(h *orgmcp.Header) map[string]any {
				builder := strings.Builder{}
				h.Render(&builder, 1)

				return map[string]any{
					"uid":        h.Uid(),
					"content":    builder.String(),
					"parent_uid": h.ParentUid(),
				}
			}),
		}

		return
	})

	if err := server.Run(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
