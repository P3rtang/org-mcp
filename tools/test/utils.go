package test

import "github.com/p3rtang/org-mcp/orgmcp"

func NewTestHeader(of *orgmcp.OrgFile) orgmcp.Uid {
	header := orgmcp.NewHeader(orgmcp.None, "Auto generated test header")

	of.AddChildren(&header)

	return header.Uid()
}

func RemoveTestHeader(of *orgmcp.OrgFile, uid orgmcp.Uid) (err error) {
	err = of.RemoveChildren(uid)

	return
}
