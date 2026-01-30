package orgmcp

import (
	"fmt"
	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/reader"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type PropValue interface {
	Date() option.Option[time.Time]
	String() string
	Int() option.Option[int]
}

type dateProperty struct {
	time time.Time
}

func (d *dateProperty) Date() option.Option[time.Time] {
	return option.Some(d.time)
}

func (d *dateProperty) String() string {
	return d.time.String()
}

func (d *dateProperty) Int() option.Option[int] {
	return option.Some(int(d.time.UnixMilli()))
}

type stringProperty struct {
	str string
}

func (s *stringProperty) Date() option.Option[time.Time] {
	return option.None[time.Time]()
}

func (s *stringProperty) String() string {
	return s.str
}

func (s *stringProperty) Int() option.Option[int] {
	if num, err := strconv.Atoi(s.str); err == nil {
		return option.Some(num)
	}

	return option.None[int]()
}

type intProperty struct {
	num int
}

func (i *intProperty) Date() option.Option[time.Time] {
	return option.None[time.Time]()
}

func (i *intProperty) String() string {
	return strconv.Itoa(i.num)
}

func (i *intProperty) Int() option.Option[int] {
	return option.Some(i.num)
}

type Properties struct {
	parent  Render
	content map[string]PropValue
}

// generateUID returns an 8-digit pseudo-random identifier as a string.
func NewPropertiesWithUID() Properties {
	return Properties{
		content: map[string]PropValue{
			"ID": &intProperty{num: rand.Intn(100000000)},
		},
	}
}

func NewPropertiesFromReader(reader *reader.PeekReader) (p Properties) {
	p.content = make(map[string]PropValue)

	bytes, err := reader.PeekBytes('\n')

	// newline not found return a default generation
	if err != nil {
		p.content["ID"] = &intProperty{num: rand.Intn(100000000)}
		return
	}
	//
	// // if the line is empty continue parsing
	// if strings.TrimSpace(string(bytes)) == "" {
	// 	_, _ = reader.ReadBytes('\n')
	// 	return NewPropertiesFromReader(reader)
	// }

	// properties not found return None
	if !strings.Contains(string(bytes), ":PROPERTIES:") {
		p.content["ID"] = &intProperty{num: rand.Intn(100000000)}
		return
	}

	// p.indent = strings.Index(string(bytes), ":PROPERTIES:")

	// advance the reader
	reader.Continue()

	// TODO: should also parse dates between [...]
	for bytes, err := reader.ReadBytes('\n'); err == nil && !strings.Contains(string(bytes), ":END:"); bytes, err = reader.ReadBytes('\n') {
		mapping := strings.SplitN(string(bytes), ":", 3)
		if len(mapping) >= 3 {
			p.content[strings.TrimSpace(mapping[1])] = &stringProperty{str: strings.TrimSpace(mapping[2])}
		}
	}

	// Assign a UID if missing
	if _, hasUID := p.content["ID"]; !hasUID {
		p.content["ID"] = &intProperty{num: rand.Intn(100000000)}
	}

	return
}

func (p *Properties) IndentLevel() int {
	return p.parent.ChildIndentLevel()
}

func (p *Properties) ChildIndentLevel() int {
	return p.IndentLevel()
}

// Render writes the properties drawer to the given strings.Builder in org-mode format.
// Example output:
// :PROPERTIES:
// :KEY1: value1
// :KEY2: value2
// :END:
func (p *Properties) Render(sb *strings.Builder) {
	if p == nil || len(p.content) == 0 {
		return
	}

	sb.WriteString(strings.Repeat(" ", p.IndentLevel()))
	sb.WriteString(":PROPERTIES:\n")

	for k, v := range p.content {
		sb.WriteString(strings.Repeat(" ", p.IndentLevel()))
		fmt.Fprintf(sb, ":%s: %s\n", k, v.String())
	}

	sb.WriteString(strings.Repeat(" ", p.IndentLevel()))
	sb.WriteString(":END:\n")
}
