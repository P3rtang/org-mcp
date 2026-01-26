package tools

import (
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/diff"
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
// It returns a diff of the changes made to the file.
func writeOrgFileToDisk(of orgmcp.OrgFile, filePath string) (res string, err error) {
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

	_, err = file.WriteString(newContent)
	if err != nil {
		return
	}

	res = diff.GetDiff(filePath, string(oldContent), newContent)

	return
}

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
