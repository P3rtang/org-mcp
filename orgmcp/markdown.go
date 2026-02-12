package orgmcp

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func (of *OrgFile) RenderMarkdown(builder *strings.Builder, depth int) {
	if depth == 0 {
		return
	}

	var headers []Render
	var content []Render

	for _, child := range of.children {
		if _, ok := child.(*Header); ok {
			headers = append(headers, child)
		} else {
			content = append(content, child)
		}
	}

	for _, child := range content {
		child.RenderMarkdown(builder, depth-1)
	}

	for _, child := range headers {
		child.RenderMarkdown(builder, depth-1)
	}
}

func (h *Header) RenderMarkdown(builder *strings.Builder, depth int) {
	if h.level > 1 {
		builder.WriteRune('\n')
	}

	builder.WriteString(strings.Repeat("#", h.level))

	builder.WriteRune(' ')
	h.status.RenderMarkdown(builder)

	builder.WriteString(orgToMarkdownStyle(h.Content))
	builder.WriteRune(' ')

	h.Progress.Then(func(p Progress) {
		p.Render(builder)
		builder.WriteRune(' ')
	})

	h.properties.RenderMarkdown(builder)

	if depth == 0 {
		return
	}

	var headers []Render
	var content []Render

	for _, child := range h.children {
		if _, ok := child.(*Header); ok {
			headers = append(headers, child)
		} else {
			content = append(content, child)
		}
	}

	for _, child := range content {
		child.RenderMarkdown(builder, depth-1)
	}

	for _, child := range headers {
		child.RenderMarkdown(builder, depth-1)
	}
}

func (b *Bullet) RenderMarkdown(builder *strings.Builder, depth int) {
	builder.WriteString(strings.Repeat(" ", b.BulletLevel()*2))

	// Render checkbox status
	switch b.checkbox {
	case NoCheck:
		fmt.Fprintf(builder, "%s ", string(b.prefix))
	case Unchecked:
		fmt.Fprintf(builder, "%s [ ] ", string(b.prefix))
	case Checked:
		fmt.Fprintf(builder, "%s [x] ", string(b.prefix))
	}

	// Render content
	builder.WriteString(orgToMarkdownStyle(b.content))
	builder.WriteRune('\n')

	if depth == 0 {
		return
	}

	for _, child := range b.children {
		child.RenderMarkdown(builder, depth-1)
	}
}

func (t *PlainText) RenderMarkdown(builder *strings.Builder, depth int) {
	builder.WriteString(strings.Repeat(" ", t.indent))
	builder.WriteString(orgToMarkdownStyle(t.content))
}

func (s HeaderStatus) RenderMarkdown(builder *strings.Builder) {
	if s == None {
		return
	}

	// check if the show-color flag has been set
	if os.Getenv("MARKDOWN_COLOR") == "" {
		builder.WriteString(s.String())
		builder.WriteRune(' ')
		return
	}

	color := "green"
	switch s {
	case Todo:
		color = "red"
	case Next:
		color = "orange"
	case Prog:
		color = "blue"
	}

	fmt.Fprintf(builder, "$\\color{%s}{\\textsf{%s}}$ ", color, s.String())
}

func orgToMarkdownStyle(content string) string {
	// font style for markdown regex conversion
	boldRegex, err := regexp.Compile(`\*(.*?)\*`)
	if err != nil {
		panic(err)
	}

	italicRegex, err := regexp.Compile("//(.*?)//")
	if err != nil {
		panic(err)
	}

	codeRegex, err := regexp.Compile("~(.*?)~")
	if err != nil {
		panic(err)
	}

	// Convert org-mode font styles to markdown
	content = codeRegex.ReplaceAllString(content, "`$1`")
	content = boldRegex.ReplaceAllString(content, "**$1**")
	content = italicRegex.ReplaceAllString(content, "*$1*")

	return content
}
