package codeblock

import (
	"strings"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/option"
)

type CodeBlock struct {
	name   option.Option[string]
	index  int
	parent Render

	lang    option.Option[string]
	content string
}

// Enforce that Bullet implements the Render interface at compile time
var _ RenderBase = (*CodeBlock)(nil)

func NewCodeBlock(content string, name option.Option[string], lang option.Option[string]) CodeBlock {
	return CodeBlock{
		name:    name,
		lang:    lang,
		content: content,
	}
}

func (c *CodeBlock) Language() option.Option[string] {
	return c.lang
}

func (c *CodeBlock) Lines() []string {
	return strings.Split(c.content, "\n")
}
