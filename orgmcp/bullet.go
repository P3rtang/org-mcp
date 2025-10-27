package orgmcp

import (
	"fmt"
	o "main/utils/option"
	"strings"
)

type bulletStatus int

const (
	NoCheck bulletStatus = iota
	Unchecked
	Checked
)

type bulletPrefix string

const (
	Star = "*"
	Dash = "-"
)

type Bullet struct {
	checkbox bulletStatus
	content  string
	location int
	prefix   bulletPrefix

	parent   o.Option[Render]
	children []Render
}

// Enforce that Bullet implements the Render interface at compile time
var _ Render = (*Bullet)(nil)

func BulletFromString(str string, parent Render) o.Option[Bullet] {
	bullet := Bullet{}

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

		bullet.content = str[5:]
	} else {
		bullet.content = str[2:]
	}

	return o.Some(bullet)
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

func (b *Bullet) Location() int {
	return b.location
}

func (b *Bullet) Render(builder *strings.Builder) {
	builder.WriteString(strings.Repeat(" ", b.IndentLevel()))

	// Render checkbox status
	switch b.checkbox {
	case NoCheck:
		fmt.Fprintf(builder, "%s ", string(b.prefix))
	case Unchecked:
		fmt.Fprintf(builder, "%s [ ]", string(b.prefix))
	case Checked:
		fmt.Fprintf(builder, "%s [X]", string(b.prefix))
	}

	// Render content
	builder.WriteString(b.content)
	builder.WriteRune('\n')
}

func (b *Bullet) IndentLevel() int {
	return o.Map(b.parent, func(r Render) int { return r.IndentLevel() }).UnwrapOr(0)
}

func (b *Bullet) AddChild(r Render) error {
	b.children = append(b.children, r)

	return nil
}

func (b *Bullet) Children() []Render {
	return b.children
}
