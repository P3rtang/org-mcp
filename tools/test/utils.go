package test

import (
	"strings"

	"github.com/p3rtang/org-mcp/orgmcp"
)

func EqualString(a, b string) bool {
	return strings.TrimSpace(a) == strings.TrimSpace(b)
}

func ContainsString(a, substr string) bool {
	return strings.Contains(a, substr)
}

func NewTestHeader(of *orgmcp.OrgFile) orgmcp.Uid {
	header := orgmcp.NewHeader(orgmcp.None, "Auto generated test header")

	of.AddChildren(&header)

	return header.Uid()
}

func RemoveTestHeader(of *orgmcp.OrgFile, uid orgmcp.Uid) (err error) {
	err = of.RemoveChildren(uid)

	return
}
