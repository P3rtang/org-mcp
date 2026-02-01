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

	parent option.Option[Render]
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
	return option.Map(p.parent, func(r Render) int {
		return r.ChildIndentLevel()
	}).UnwrapOr(0)
}

func (p *PlainText) ChildIndentLevel() int {
	return p.IndentLevel() + 2
}

func (p *PlainText) Level() int {
	return option.Map(p.parent, func(r Render) int {
		return r.Level() + 1
	}).UnwrapOr(0)
}

func (p *PlainText) Location(table map[Uid]int) (loc int) {
	if val, ok := table[p.Uid()]; ok {
		return val
	}

	if parent, ok := p.parent.Split(); ok {
		loc += parent.Location(table)

		for i, child := range parent.Children() {
			if child.Uid() == p.Uid() {
				loc += i + 1
				break
			}
		}
	}

	return
}

func (p *PlainText) AddChildren(r ...Render) error {
	return errors.New("PlainText cannot have children")
}

func (p *PlainText) SetParent(r Render) error {
	p.parent = option.Some(r)
	p.index = len(r.Children())

	return nil
}

func (p *PlainText) RemoveChildren(...Uid) error {
	return errors.New("PlainText cannot have children")
}

func (p *PlainText) Children() []Render {
	return []Render{}
}

func (p *PlainText) ChildrenRec(_ int) []Render {
	return []Render{}
}

func (p *PlainText) Uid() Uid {
	if p.parent.IsNone() {
		return NewUid(-1)
	}

	return NewUid(fmt.Sprintf("%s.t%d", p.parent.Unwrap().Uid(), p.index))
}
func (p *PlainText) ParentUid() Uid {
	if p.parent.IsNone() {
		return NewUid(0)
	}
	return p.parent.Unwrap().Uid()
}

func (p *PlainText) Status() HeaderStatus {
	return option.Map(p.parent, func(p Render) HeaderStatus {
		return p.Status()
	}).UnwrapOr(None)
}

func (p *PlainText) TagList() (list TagList) {
	if parent, ok := p.parent.Split(); ok {
		list = parent.TagList()
	}

	return
}

func (p *PlainText) Preview() string {
	return p.content
}

func (p *PlainText) Path() string {
	if parent, ok := p.parent.Split(); ok {
		return parent.Path() + "/" + p.Uid().String()
	}

	return p.Uid().String()
}
