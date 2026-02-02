package tools

import (
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
)

type ViewItem struct {
	Uid     string               `json:"uid,omitempty" jsonschema:"description=UID of the header to view. If not provided, all headers are considered."`
	Status  *orgmcp.RenderStatus `json:"status,omitempty" jsonschema:"description=Filter headers by status (e.g. TODO | DONE). Case insensitive. As well as bullets by their checkbox status (e.g. CHECKED | UNCHECKED)."`
	Content string               `json:"content,omitempty" jsonschema:"description=Filter headers with a regex match on content. It will only consider the preview of the header content and not any metadata; children; status or other information."`
	Tags    []string             `json:"tags,omitempty" jsonschema:"description=Filter headers by tags. Only headers containing all specified tags will be returned."`
	Depth   *int                 `json:"depth,omitempty" jsonschema:"description=Depth of child headers to include. Default is 1 (only direct children)."`
}

type ViewInput struct {
	Items   []ViewItem       `json:"items" jsonschema:"description=List of items to view based on their UIDs and filters."`
	Columns []*orgmcp.Column `json:"columns,omitempty" jsonschema:"description=List of columns to include in the output. If not specified, defaults to [UID, PREVIEW]. Always prefer preview over content to reduce output size, any metadata can be fetched with additional columns. Only use content if the rendered output that the user sees is important."`
	Path    string           `json:"path,omitempty" jsonschema:"description=An optional file path, will default to the configured org file. (./.tasks.org)"`
}

var ViewTool = mcp.Tool{
	Name: "query_items",
	Description: "View headers, bullets, or other items in the org file based on their UIDs and filters like status and tags. You can specify multiple items to view in a single call. " +
		"Consider using the 'depth' parameter to include child items in the output. The results will be returned in CSV format with the specified columns. " +
		"Each item in the 'items' array functions as an OR operation between items, but an AND operation inside a single item. " +
		"The function can and will return multiple items if they match the criteria specified. ",
	InputSchema: mcp.GenerateSchema(ViewInput{}),
	Callback: func(args map[string]any, options mcp.FuncOptions) (resp []any, err error) {
		bytes, err := json.Marshal(args)
		if err != nil {
			return nil, fmt.Errorf("error marshalling header input: %v", err)
		}

		var input ViewInput
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

		if len(input.Columns) == 0 {
			uidCol := orgmcp.ColUid
			contentCol := orgmcp.ColContent
			input.Columns = []*orgmcp.Column{&uidCol, &contentCol}
		}

		results := map[orgmcp.Uid]orgmcp.Render{}

		for _, item := range input.Items {
			depth := 1
			if item.Depth != nil {
				depth = *item.Depth
			}

			for _, render := range orgFile.ChildrenRec(-1) {
				if item.Uid != "" && render.Uid().String() != item.Uid {
					continue
				}

				if item.Status != nil && render.Status() != *item.Status {
					continue
				}

				if item.Content != "" {
					reg, err := regexp.Compile(item.Content)
					if err != nil {
						return nil, err
					}

					if !reg.MatchString(render.Preview(-1)) {
						continue
					}
				}

				if len(item.Tags) > 0 {
					foundAll := true
					for _, tag := range item.Tags {
						if !slices.Contains(render.TagList(), tag) {
							foundAll = false
							break
						}
					}

					if !foundAll {
						continue
					}
				}

				results[render.Uid()] = render
				for _, child := range render.ChildrenRec(depth) {
					results[child.Uid()] = child
				}
			}
		}

		locationTable := orgFile.GetLocationTable()

		ordered := slices.Collect(maps.Values(results))
		slices.SortFunc(ordered, func(a, b orgmcp.Render) int {
			return (*locationTable)[a.Uid()] - (*locationTable)[b.Uid()]
		})

		resp = append(resp, orgmcp.PrintCsv(ordered, input.Columns))

		return
	},
}
