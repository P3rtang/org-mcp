package codeblock

import (
	"fmt"
	"strings"
)

func (c *CodeBlock) Render(builder *strings.Builder, depth int) {
	indent := strings.Repeat(" ", c.IndentLevel())

	c.name.Then(func(s string) { fmt.Fprintf(builder, "%s#+NAME: %s\n", indent, s) })

	fmt.Fprintf(builder, "%s#+BEGIN_SRC", indent)

	c.lang.Then(func(s string) {
		builder.WriteRune(' ')
		builder.WriteString(s)
	})

	builder.WriteRune('\n')

	for line := range strings.Lines(c.content) {
		fmt.Fprintf(builder, "%s%s\n", indent, strings.TrimRight(line, "\n"))
	}

	fmt.Fprintf(builder, "%s#+END_SRC\n", indent)
}

func (c *CodeBlock) RenderMarkdown(builder *strings.Builder, depth int) {
	indent := strings.Repeat(" ", c.IndentLevel())

	fmt.Fprintf(builder, "%s```%s\n", indent, c.lang.UnwrapOrDefault())
	fmt.Fprintf(builder, "%s%s", indent, c.content)
	fmt.Fprintf(builder, "%s```\n", indent)
}

func (c *CodeBlock) Preview(length int) string {
	if length < 0 {
		return c.content
	}

	lines := c.Lines()

	if len(lines) <= 1 {
		return fmt.Sprintf("EMPTY CODE BLOCK ; LANG=`%s`", c.Language().UnwrapOr("UNKNOWN"))
	}

	return fmt.Sprintf("CODE BLOCK ; LANG=`%s`; FIRST LINE: %s", c.Language().UnwrapOr("UNKNOWN"), c.Lines()[0])
}
