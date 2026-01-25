package orgmcp

import (
	"errors"
	"fmt"
	"github.com/p3rtang/org-mcp/utils/option"
	o "github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/reader"
	"github.com/p3rtang/org-mcp/utils/slice"
	"os"
	"slices"
	"strings"
)

type bulletStatus int

const (
	NoCheck bulletStatus = iota
	Unchecked
	Checked
)

func NewBulletStatus(str string) bulletStatus {
	switch strings.ToLower(str) {
	case "unchecked":
		return Unchecked
	case "checked":
		fallthrough
	case "check":
		return Checked
	default:
		return NoCheck
	}
}

type bulletPrefix string

const (
	Star bulletPrefix = "*"
	Dash bulletPrefix = "-"
)

type Bullet struct {
	checkbox bulletStatus
	content  string
	prefix   bulletPrefix
	index    int
	indent   int

	parent   o.Option[Render]
	children []Render
}

// Enforce that Bullet implements the Render interface at compile time
var _ Render = (*Bullet)(nil)

func NewBulletFromReader(r *reader.PeekReader) o.Option[*Bullet] {
	bullet := Bullet{}

	line, err := r.ReadBytes('\n')
	fmt.Fprintf(os.Stderr, "Parsing bullet: %s\n", string(line))
	if err != nil {
		return o.None[*Bullet]()
	}

	str := strings.TrimLeft(string(line), " ")
	bullet.indent = len(line) - len(str)

	if len(str) < 2 {
		return o.None[*Bullet]()
	}

	switch str[0] {
	case '*':
		bullet.prefix = Star
	case '-':
		bullet.prefix = Dash
	default:
		return o.None[*Bullet]()
	}

	if len(str) > 4 && str[2] == '[' && str[4] == ']' {
		bullet.checkbox = Unchecked

		switch str[3] {
		case 'X':
			fallthrough
		case 'x':
			bullet.checkbox = Checked
		}

		bullet.content = strings.TrimSpace(str[5:])
	} else {
		bullet.content = strings.TrimSpace(str[2:])
	}

	for line, err := r.PeekBytes('\n'); err == nil; line, err = r.PeekBytes('\n') {
		if len(strings.TrimSpace(string(line))) == 0 ||
			strings.TrimSpace(string(line))[0] == '*' ||
			strings.TrimSpace(string(line))[0] == '-' {
			break
		}

		fmt.Fprintf(os.Stderr, "Appending to bullet content: %s\n", string(line))

		r.Continue()
		bullet.content += "\n" + strings.TrimRight(string(line), "\n")
	}

	return option.Some(&bullet)
}

func NewBulletFromString(str string, parent Render) o.Option[Bullet] {
	bullet := Bullet{}

	if parent != nil {
		bullet.index = len(parent.Children())
	}

	// fmt.Fprintf(os.Stderr, "%s\n", str)

	if parent == nil {
		bullet.parent = o.None[Render]()
	} else {
		bullet.parent = o.Some(parent)
	}

	if str[1] != ' ' {
		return o.None[Bullet]()
	}

	switch str[0] {
	case '*':
		bullet.prefix = Star
	case '-':
		bullet.prefix = Dash
	default:
		panic("unreachable")
	}

	if str[2] == '[' && str[4] == ']' {
		bullet.checkbox = Unchecked

		switch str[3] {
		case 'X':
			fallthrough
		case 'x':
			bullet.checkbox = Checked
		}

		bullet.content = strings.TrimSpace(str[5:])
	} else {
		bullet.content = strings.TrimSpace(str[2:])
	}

	return o.Some(bullet)
}

func NewBullet(parent Render, status bulletStatus) *Bullet {
	prefix := Dash
	if status == NoCheck {
		prefix = Star
	}

	bullet := &Bullet{
		index:    len(parent.Children()),
		parent:   o.Some(parent),
		prefix:   prefix,
		checkbox: status,
	}

	bullet.indent = parent.IndentLevel()
	parent.AddChildren(bullet)

	return bullet
}

func (b *Bullet) SetIndex(idx int) {
	b.index = idx
}

func (b *Bullet) SetContent(content string) {
	b.content = content
}

func (b *Bullet) CheckProgress() o.Option[Progress] {
	switch b.checkbox {
	case NoCheck:
		return o.None[Progress]()
	case Unchecked:
		return o.Some(Progress{done: o.Some(false)})
	case Checked:
		return o.Some(Progress{done: o.Some(true)})
	default:
		panic("unreachable")
	}
}

func (b *Bullet) Render(builder *strings.Builder, depth int) {
	builder.WriteString(strings.Repeat(" ", b.IndentLevel()))

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
	builder.WriteString(b.content)
	builder.WriteRune('\n')

	if depth == 0 {
		return
	}

	for _, child := range b.children {
		child.Render(builder, depth-1)
	}
}

func (b *Bullet) IndentLevel() int {
	return b.indent
}

func (b *Bullet) AddChildren(r ...Render) error {
	for _, child := range r {
		if _, ok := child.(*Bullet); !ok {
			return errors.New("can only add Bullet children to Bullet")
		}

		child.SetParent(b)
	}

	b.children = append(b.children, r...)

	return nil
}

func (b *Bullet) SetParent(render Render) error {
	if option.Map(b.parent, func(r Render) Uid { return r.Uid() }) == option.Some(render.Uid()) {
		return nil
	}

	b.index = len(render.Children())
	b.parent = option.Some(render)

	return nil
}

func (b *Bullet) RemoveChildren(uids ...Uid) error {
	b.children = slice.Filter(b.children, func(r Render) bool {
		return slices.Contains(uids, b.Uid())
	})

	return nil
}

func (b *Bullet) Children() []Render {
	return b.children
}

func (b *Bullet) ChildrenRec() []Render {
	children := []Render{}

	for _, child := range b.Children() {
		children = append(children, child.ChildrenRec()...)
	}

	return children
}

func (b *Bullet) Uid() Uid {
	return option.Map(b.parent, func(r Render) Uid {
		return NewUid(fmt.Sprintf("%s.b%d", r.Uid(), b.index))
	}).UnwrapOr(NewUid(-1))
}

func (b *Bullet) ParentUid() Uid {
	return option.Map(b.parent, func(r Render) Uid {
		return r.Uid()
	}).UnwrapOr(NewUid(0))
}

// ToggleCheckbox toggles the checkbox state between Unchecked and Checked
// Only works for bullets that already have a checkbox (not NoCheck)
func (b *Bullet) ToggleCheckbox() {
	switch b.checkbox {
	case Unchecked:
		b.checkbox = Checked
	case Checked:
		b.checkbox = Unchecked
	}
}

// CompleteCheckbox marks the checkbox as completed (checked)
// Only works for bullets that already have a checkbox (not NoCheck)
func (b *Bullet) CompleteCheckbox() {
	if b.checkbox != NoCheck {
		b.checkbox = Checked
	}
}

// HasCheckbox returns true if the bullet has a checkbox
func (b *Bullet) HasCheckbox() bool {
	return b.checkbox != NoCheck
}
