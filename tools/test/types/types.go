package types

import "github.com/p3rtang/org-mcp/tools"

type ManageHeaderTest struct {
	name     string
	input    tools.HeaderInput
	expected []any
}

type ManageBulletTest struct {
	name     string
	input    tools.BulletInput
	expected []any
}

type ManageTextTest struct {
	name        string
	input       tools.TextInputSchema
	expected    []any
	expectEmpty bool
}

type TableTest struct {
	Name        string
	Input       tools.QueryTableInput
	Expected    []any
	ExpectEmpty bool
}

type CodeBlockTest struct{}
