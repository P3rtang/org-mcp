package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"main/mcp"
	"main/orgmcp"
	"main/utils/slice"
	"os"
	"strings"
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
	// Define command-line flags
	workspace := flag.String("workspace", "", "Path to the current workspace, when using relative file paths this is the root directory.")
	flag.Parse()

	defaultPath := ""

	if *workspace != "" {
		defaultPath += *workspace + "/"
	}

	defaultPath += ".tasks.org"

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
		func(args map[string]any) (resp any, err error) {
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
				h.Render(&builder, 0)

				return map[string]any{
					"uid":        h.Uid(),
					"content":    builder.String(),
					"parent_uid": h.ParentUid(),
				}
			})

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
				"You should think of headers as a directory containing other dirs (sub-headers), but also files and metadata.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"index": map[string]any{
						"type":        "number",
						"description": "The index of the header to view.",
					},
					"depth": map[string]any{
						"type":        "number",
						"description": "The depth to which children should be rendered.",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "An optional file path, will default to the configured org file.",
					},
				},
			},
		}, func(args map[string]any) (resp any, err error) {
			filePath, ok := args["path"].(string)

			if !ok || filePath == "" {
				filePath = defaultPath
			}

			of, err := loadOrgFile(filePath)
			if err != nil {
				return
			}

			// Get the index (uid) from the arguments
			var uid int
			if indexVal, ok := args["index"].(float64); ok {
				uid = int(indexVal)
			} else {
				return nil, errors.New("invalid or missing index parameter")
			}

			// Get the depth parameter, default to 1 if not provided
			depth := 1
			if depthVal, ok := args["depth"].(float64); ok {
				depth = int(depthVal)
			}

			// Retrieve the header by uid
			header, ok := of.GetUid(uid).Split()
			if !ok {
				return nil, fmt.Errorf("header with uid %d not found", uid)
			}

			// Render the header with the specified depth
			builder := strings.Builder{}
			header.Render(&builder, depth)

			response := map[string]any{
				"uid":        header.Uid(),
				"content":    builder.String(),
				"parent_uid": header.ParentUid(),
			}

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
		}, func(args map[string]any) (resp any, err error) {
			response := []any{}
			uids := []int{}

			if idcs, ok := args["uid"].([]any); ok {
				for _, idx := range idcs {
					if i, ok := idx.(float64); ok {
						uids = append(uids, int(i))
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
				if r, ok = org_file.GetUid(uid).Split(); !ok {
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
		func(args map[string]any) (resp any, err error) {
			filePath, ok := args["path"].(string)
			if !ok || filePath == "" {
				filePath = defaultPath
			}

			of, err := loadOrgFile(filePath)
			if err != nil {
				return
			}

			parentUidFloat, ok := args["parent_uid"].(float64)
			if !ok {
				return nil, errors.New("Invalid or missing parent_uid parameter")
			}
			parentUid := int(parentUidFloat)

			title, ok := args["title"].(string)
			if !ok || strings.TrimSpace(title) == "" {
				return nil, errors.New("Title must not be empty.")
			}

			statusStr, _ := args["status"].(string)

			status := orgmcp.StatusFromString(statusStr)
			if status == orgmcp.None {
				status = orgmcp.Todo
			}

			render, ok := of.GetUid(parentUid).Split()
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

	server.AddTool(
		mcp.Tool{
			Name: "complete_checkbox",
			Description: "Toggle or complete a checkbox in a bullet point within a header.\n" +
				"Specify the action to either 'toggle' the checkbox state or 'complete' to mark it as done.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []string{"header_uid", "bullet_index", "action"},
				"properties": map[string]any{
					"header_uid": map[string]any{
						"type":        "number",
						"description": "The UID of the header containing the bullet with the checkbox.",
					},
					"bullet_index": map[string]any{
						"type":        "number",
						"description": "The index (0-based) of the bullet within the header's children.",
					},
					"action": map[string]any{
						"type":        "string",
						"enum":        []string{"toggle", "complete"},
						"description": "Action to perform: 'toggle' to toggle between checked/unchecked, or 'complete' to mark as checked.",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "An optional file path, will default to the configured org file.",
					},
				},
			},
		},
		func(args map[string]any) (resp any, err error) {
			filePath, ok := args["path"].(string)
			if !ok || filePath == "" {
				filePath = defaultPath
			}

			org_file, err := loadOrgFile(filePath)
			if err != nil {
				return
			}

			headerUidFloat, ok := args["header_uid"].(float64)
			if !ok {
				return nil, errors.New("invalid or missing header_uid parameter")
			}
			headerUid := int(headerUidFloat)

			bulletIndexFloat, ok := args["bullet_index"].(float64)
			if !ok {
				return nil, errors.New("invalid or missing bullet_index parameter")
			}
			bulletIndex := int(bulletIndexFloat)

			action, ok := args["action"].(string)
			if !ok {
				return nil, errors.New("invalid or missing action parameter")
			}

			if action != "toggle" && action != "complete" {
				return nil, errors.New("action must be either 'toggle' or 'complete'")
			}

			header, ok := org_file.GetUid(headerUid).Split()
			if !ok {
				return nil, fmt.Errorf("header with uid %d not found", headerUid)
			}

			var bullet *orgmcp.Bullet
			if action == "toggle" {
				bullet, err = header.ToggleCheckboxByIndex(bulletIndex)
			} else {
				bullet, err = header.CompleteCheckboxByIndex(bulletIndex)
			}

			if err != nil {
				return
			}

			builder := strings.Builder{}
			org_file.Render(&builder, -1)

			file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			if err != nil {
				return
			}
			defer file.Close()

			file.WriteString(builder.String())

			builder.Reset()
			bullet.Render(&builder, 0)

			resp = map[string]any{
				"header_uid":   headerUid,
				"bullet_index": bulletIndex,
				"content":      builder.String(),
				"action":       action,
			}

			return
		},
	)

	server.AddTool(
		mcp.Tool{
			Name:        "bullet_point",
			Description: "Add or remove a bullet point from a header/task. Use method 'add' to insert a bullet point, or 'remove' to delete one. Index is used for both cases.",
			InputSchema: map[string]any{
				"type":     "object",
				"required": []string{"method", "header_uid", "index"},
				"properties": map[string]any{
					"method": map[string]any{
						"type":        "string",
						"enum":        []string{"add", "remove"},
						"description": "Operation to perform: 'add' to insert a bullet, 'remove' to delete a bullet.",
					},
					"header_uid": map[string]any{
						"type":        "number",
						"description": "UID of the header/task to modify.",
					},
					"index": map[string]any{
						"type":        "number",
						"description": "The index at which to add or remove the bullet.",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Text content of the bullet (required for add).",
					},
					"checkbox": map[string]any{
						"type": "string",
						"description": "The current status of a checkbox." +
							"For a bullet without a checkbox, you should omit this parameter",
						// TODO: add partial checked
						"enum": []string{"Unchecked", "Checked"},
					},
					"path": map[string]any{
						"type":        "string",
						"description": "Optional file path, defaults to .tasks.org.",
					},
				},
			},
		},
		func(args map[string]any) (resp any, err error) {
			filePath, ok := args["path"].(string)
			if !ok || filePath == "" {
				filePath = defaultPath
			}

			of, err := loadOrgFile(filePath)
			if err != nil {
				return nil, err
			}

			headerUidFloat, ok := args["header_uid"].(float64)
			if !ok {
				return nil, errors.New("invalid or missing header_uid parameter")
			}
			headerUid := int(headerUidFloat)

			method, ok := args["method"].(string)
			if !ok {
				return nil, errors.New("invalid or missing method parameter")
			}

			indexFloat, ok := args["index"].(float64)
			if !ok {
				return nil, errors.New("invalid or missing index parameter")
			}
			index := int(indexFloat)

			header, ok := of.GetUid(headerUid).Split()
			if !ok {
				return nil, fmt.Errorf("header with uid %d not found", headerUid)
			}

			if method == "add" {
				content, ok := args["content"].(string)
				if !ok || strings.TrimSpace(content) == "" {
					return nil, errors.New("content is required for add method")
				}

				var c string
				if c, ok = args["checkbox"].(string); !ok {
					return nil, errors.New("invalid or missing checkbox parameter")
				}

				bullet := orgmcp.NewBullet(header, orgmcp.NewBulletStatus(c))
				header.AddChildren(&bullet)

				builder := strings.Builder{}
				of.Render(&builder, -1)

				file, ferr := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
				if ferr != nil {
					return nil, ferr
				}
				defer file.Close()

				file.WriteString(builder.String())

				builder.Reset()
				bullet.Render(&builder, 0)

				resp = map[string]any{
					"header_uid": headerUid,
					"index":      index,
					"content":    builder.String(),
					"method":     method,
				}
				return resp, nil
			} else if method == "remove" {
			}

			return nil, errors.New("invalid method; must be 'add' or 'remove'")
		},
	)

	if err := server.Run(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
