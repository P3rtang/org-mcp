package orgmcp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/reader"
)

type PlainText struct {
	content string
	indent  int
	index   int

	parent Render
}

// Enforce that PlainText implements the Render interface at compile time
var _ Render = (*PlainText)(nil)

func NewPlainTextFromReader(reader *reader.PeekReader) option.Option[*PlainText] {
	line, err := reader.ReadBytes('\n')

	if err != nil {
		return option.None[*PlainText]()
	}

	content := strings.TrimLeft(string(line), " ")
	indent := len(line) - len(content)

	content = strings.TrimSpace(content) + "\n"

	return option.Some(&PlainText{content: content, indent: indent})
}

func (p *PlainText) CheckProgress() option.Option[Progress] {
	return option.None[Progress]()
}

func (p *PlainText) Render(builder *strings.Builder, depth int) {
	builder.WriteString(strings.Repeat(" ", p.indent))
	builder.WriteString(p.content)
}

func (p *PlainText) IndentLevel() int {
	return p.indent
}

func (p *PlainText) AddChildren(r ...Render) error {
	return errors.New("PlainText cannot have children")
}

func (p *PlainText) SetParent(r Render) error {
	p.parent = r
	p.index = len(r.Children())

	return nil
}

func (p *PlainText) RemoveChildren(...Uid) error {
	return errors.New("PlainText cannot have children")
}

func (p *PlainText) Children() []Render {
	return []Render{}
}

func (p *PlainText) ChildrenRec() []Render {
	return []Render{}
}

func (p *PlainText) Uid() Uid {
	if p.parent == nil {
		return NewUid(-1)
	}

	return NewUid(fmt.Sprintf("%s.t%d", p.parent.Uid(), p.index))
}
func (p *PlainText) ParentUid() Uid {
	if p.parent == nil {
		return NewUid(0)
	}
	return p.parent.Uid()
}
