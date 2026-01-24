package tools

import (
	"os"
	"strings"

	"github.com/p3rtang/org-mcp/orgmcp"
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
