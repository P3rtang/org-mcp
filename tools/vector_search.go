package tools

import (
	"context"
	"slices"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
)

type VectorSearchInput struct {
	Query   string           `json:"query" jsonschema:"description=The search query string.,required=true"`
	TopN    int              `json:"top_n,omitempty" jsonschema:"description=The number of top relevant headers to return."`
	Path    string           `json:"path,omitempty" jsonschema:"description=An optional file path; will default to the configured org file. (./.tasks.org)"`
	Columns []*orgmcp.Column `json:"columns,omitempty" jsonschema:"description=List of columns to include in the output. If not specified defaults to [UID | PREVIEW]."`
}

var VectorSearch = mcp.GenericTool[VectorSearchInput]{
	Name: "vector_search",
	Description: "Perform a vector search on all headers in the org file based on the provided query string. " +
		"Returns the top N most relevant headers, bullet, text or other element.\n" +
		"It is optimal to include as much information as possible in the query, overflowing the text limit is hard. " +
		"So include context like timeframes, people involved, locations, etc. to get the best results.",
	Callback: func(ctx context.Context, input VectorSearchInput, options mcp.FuncOptions) (resp []any, err error) {
		filePath := options.DefaultPath
		if input.Path != "" {
			filePath = input.Path
		}

		if input.TopN <= 0 {
			input.TopN = 3.0
		}

		of, err := mcp.LoadOrgFile(ctx, filePath)
		if err != nil {
			return
		}

		locationTable := of.BuildLocationTable()
		searchResults, err := of.VectorSearch(input.Query, input.TopN)

		slices.SortFunc(searchResults, func(a, b orgmcp.Render) int {
			return (*locationTable)[a.Uid()] - (*locationTable)[b.Uid()]
		})

		resp = append(resp, orgmcp.PrintCsv(searchResults, input.Columns))

		return
	},
}
