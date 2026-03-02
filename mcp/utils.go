package mcp

import (
	"context"
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/diff"
)

// LoadOrgFile loads an OrgFile from the given file path.
// It opens the file, reads it using OrgFileFromReader, and returns the result.
func LoadOrgFile(ctx context.Context, filePath string) (*orgmcp.OrgFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	of, err := orgmcp.OrgFileFromReader(ctx, file).Split()
	context.WithValue(ctx, orgmcp.ORG_FILE_KEY, of)

	return of, err
}

// writeOrgFileToDisk renders the OrgFile and writes it to the provided file path.
// It returns a diff of the changes made to the file.
func WriteOrgFileToDisk(ctx context.Context, of *orgmcp.OrgFile, filePath string) (res string, err error) {
	oldContent, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	builder := strings.Builder{}
	of.Render(&builder, -1)
	newContent := builder.String()

	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}

	_, err = file.WriteString(newContent)
	if err != nil {
		return
	}

	res = diff.GetDiff(filePath, string(oldContent), newContent)

	return
}
