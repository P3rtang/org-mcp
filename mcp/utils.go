package mcp

import (
	"os"

	"github.com/p3rtang/org-mcp/orgmcp"
)

// LoadOrgFile loads an OrgFile from the given file path.
// It opens the file, reads it using OrgFileFromReader, and returns the result.
func LoadOrgFile(filePath string) (orgmcp.OrgFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return orgmcp.OrgFile{}, err
	}
	defer file.Close()

	return orgmcp.OrgFileFromReader(file).Split()
}
