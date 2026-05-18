package test

import (
	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/tools"
)

func intoOneOfArray[T tools.QueryTableSimple | tools.QueryTableUnion](t ...T) []mcp.GenericOneOf[*tools.QueryTableUnion, tools.ApplicableTool] {
	entries := []mcp.GenericOneOf[*tools.QueryTableUnion, tools.ApplicableTool]{}

	for _, input := range t {
		entries = append(entries, mcp.GenericOneOf[*tools.QueryTableUnion, tools.ApplicableTool]{
			Value: tools.NewTableInputUnion(input),
		})
	}

	return entries
}
