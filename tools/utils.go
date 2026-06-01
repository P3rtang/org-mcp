package tools

import (
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/diff"
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
	affectedItems map[orgmcp.Uid]orgmcp.Render
	err           error
}
