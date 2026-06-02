package tools

import (
	"context"
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/orgmcp"
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/diff"
)

const (
	ITEM_NOT_FOUND = "Uid %s was not found."
	WRONG_TYPE     = "Uid %s is not a %s, but %s."
	EMPTY_CONTEXT  = "%s was not found in context. This is a bug, please report."
)

// GetDiffOnly renders the OrgFile and returns a diff against the current disk content
// without modifying the file on disk.
func GetDiffOnly(of orgmcp.OrgFile, filePath string) (res string, err error) {
	oldContent, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	builder := strings.Builder{}
	of.Render(&builder, -1)
	newContent := builder.String()

	res = diff.GetDiff(filePath, string(oldContent), newContent)

	return
}

type ApplyResult struct {
	affectedItems map[Uid]Render
	err           error
}

type ApplicableTool interface {
	Apply(ctx context.Context) (ApplyResult, error)
}

type TableApplicableTool interface {
	Apply(ctx context.Context) TableApplyResult
}
