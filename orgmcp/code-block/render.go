package codeblock

import (
	"fmt"
	"strings"
)

func (c *CodeBlock) Render(builder *strings.Builder, depth int) {
	panic("NOT IMPLEMENTED YET")
}

func (c *CodeBlock) RenderMarkdown(builder *strings.Builder, depth int) {
	panic("NOT IMPLEMENTED YET")
}

func (c *CodeBlock) Preview(length int) string {
	lines := c.Lines()

	if len(lines) == 0 {
		return fmt.Sprintf("EMPTY CODE BLOCK ; LANG=`%s`", c.Language().UnwrapOr("UNKNOWN"))
	}

	return fmt.Sprintf("CODE BLOCK ; LANG=`%s`; FIRST LINE: %s", c.Language().UnwrapOr("UNKNOWN"), c.Lines()[0])
}
