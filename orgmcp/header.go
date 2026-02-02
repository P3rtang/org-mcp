package orgmcp

import (
	"errors"
	"fmt"
	"iter"
	"os"
	"slices"
	"strings"

	"github.com/p3rtang/org-mcp/embeddings"
	"github.com/p3rtang/org-mcp/utils/itertools"
	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/reader"
	"github.com/p3rtang/org-mcp/utils/slice"
)

type HeaderStatus string

func StatusFromString(str string) HeaderStatus {
	switch strings.ToLower(str) {
	case "todo":
		return Todo
	case "next":
		return Next
	case "prog":
		return Prog
	case "revw":
		return Revw
	case "done":
		return Done
	case "delg":
		return Delg
	}

	return None
}

func (h *HeaderStatus) UnmarshalJSON(input []byte) error {
	switch strings.Trim(strings.ToLower(string(input)), "\"") {
	case "todo":
		fmt.Fprintf(os.Stderr, "Found todo")
		*h = Todo
	case "next":
		*h = Next
	case "prog":
		*h = Prog
	case "revw":
		*h = Revw
	case "done":
		*h = Done
	case "delg":
		*h = Delg
	case "none":
	case "":
		*h = None
	default:
		return errors.New("invalid HeaderStatus value")
	}

	return nil
}

func (s HeaderStatus) String() string {
	if s == None {
		return ""
	}

	return string(s)
}

func (s HeaderStatus) GetNext() HeaderStatus {
	switch s {
	case None:
		return Todo
	case Todo:
		return Next
	case Next:
		return Prog
	case Prog:
		return Done
	case Done:
		return None
	default:
		panic("unreachable")
	}
}

const (
	None HeaderStatus = "NONE"
	Todo HeaderStatus = "TODO"
	Next HeaderStatus = "NEXT"
	Prog HeaderStatus = "PROG"
	Revw HeaderStatus = "REVW"
	Done HeaderStatus = "DONE"
	Delg HeaderStatus = "DELG"
)

var SPECIAL_TOKENS = []string{"[", ":"}

// Enforce that Header implements the Render interface at compile time
var _ Render = (*Header)(nil)

type Header struct {
	status   HeaderStatus
	level    int
	Progress option.Option[Progress]
	Tags     option.Option[TagList]
	location int

	Parent     option.Option[Render]
	children   []Render
	schedule   option.Option[Schedule]
	properties Properties
	embedding  option.Option[embeddings.Embedding]

	Content string
}

func NewHeader(status HeaderStatus, content string) Header {
	header := Header{
		status:   status,
		Progress: option.None[Progress](),
		Tags:     option.None[TagList](),

		Parent:    option.None[Render](),
		children:  []Render{},
		schedule:  option.None[Schedule](),
		embedding: option.None[embeddings.Embedding](),

		Content: content,
	}

	header.properties = NewPropertiesWithUID(&header)

	return header
}

// TODO: remove str from arguments and parse from reader only (prob make a new constructor)
func NewHeaderFromString(str string, reader *reader.PeekReader) option.Option[Header] {
	// fmt.Fprintf(os.Stderr, "Parsing header %s\n", str)
	if !strings.HasPrefix(str, "*") {
		return option.None[Header]()
	}

	// Trim the line to remove any trailing newline characters
	str = strings.TrimSpace(str)

	header := Header{}
	next, stop := iter.Pull(strings.SplitSeq(str, " "))
	defer stop()

	level, _ := next()
	header.level = strings.Count(level, "*")

	part, end := next()

	header.status = StatusFromString(part)

	if header.status != None {
		part, end = next()
	}

	for end {
		if slice.Any(SPECIAL_TOKENS, func(char string) bool { return strings.HasPrefix(part, char) }) {
			switch part[0] {
			case '[':
				header.Progress = ProgressFromString(part)
			case ':':
				header.Tags = TagListFromString(part)
			}
		} else {
			header.Content += part
			header.Content += " "
		}

		part, end = next()
		continue
	}

	header.Content = strings.TrimSpace(header.Content)

	if reader == nil {
		return option.Some(header)
	}

	header.schedule = option.Map(NewScheduleFromReader(reader), func(s Schedule) Schedule {
		s.parent = &header
		return s
	})

	header.properties = NewPropertiesFromReader(reader)
	header.properties.parent = &header

	return option.Some(header)
}

func (h *Header) AddChildren(render ...Render) error {
	for _, child := range render {
		child.SetParent(h)
	}

	h.children = append(h.children, render...)

	return nil
}

func (h *Header) SetParent(render Render) error {
	h.Parent = option.Some(render)
	h.level = render.Level() + 1
	h.location = len(render.Children())

	return nil
}

func (b *Header) RemoveChildren(uids ...Uid) error {
	b.children = slice.Filter(b.children, func(r Render) bool {
		return !slices.Contains(uids, r.Uid())
	})

	return nil
}

func (h *Header) Render(builder *strings.Builder, depth int) {
	builder.WriteString(strings.Repeat("*", h.Level()))
	builder.WriteString(" ")
	if h.status != None {
		builder.WriteString(h.status.String())
		builder.WriteString(" ")
	}
	builder.WriteString(h.Content)

	h.Progress.Then(func(p Progress) {
		builder.WriteRune(' ')
		p.Render(builder)
	})

	h.Tags.Then(func(tl TagList) {
		builder.WriteRune(' ')
		tl.Render(builder)
	})

	if depth == 0 {
		builder.WriteString("...")
		builder.WriteRune('\n')
		return
	}

	builder.WriteRune('\n')

	h.schedule.Then(func(s Schedule) {
		s.Render(builder)
	})

	h.properties.Render(builder)

	children := itertools.FromSlice(h.children)
	itertools.ForEach(children, func(child Render) {
		child.Render(builder, depth-1)
	})
}

func (h *Header) CheckProgress() option.Option[Progress] {
	if h.Progress.IsNone() && h.status != None {
		return option.Some(Progress{done: option.Some(h.status == Done)})
	}

	return option.Map(h.Progress, func(_ Progress) Progress {
		progress := Progress{}

		for _, child := range h.children {
			child_progress := child.CheckProgress()

			if child_progress.IsNone() {
				continue
			}

			if child_progress.UnwrapPtr().Done() {
				progress.Complete += 1
			}

			progress.Total += 1
		}

		h.Progress = option.Some(progress)

		if h.Progress.AndThen(func(p Progress) bool { return p.Done() }) && h.status != None {
			h.status = Done
		} else if h.Progress.AndThen(func(p Progress) bool { return p.Prog() }) && h.status != None && h.status != Done {
			h.status = Prog
		}

		return progress
	})
}

func (p *Header) Location(table map[Uid]int) (loc int) {
	if val, ok := table[p.Uid()]; ok {
		return val
	}

	if parent, ok := p.Parent.Split(); ok {
		loc += parent.Location(table)

		for i, child := range parent.ChildrenRec(-1) {
			if child.Uid() == p.Uid() {
				loc += i + 1
				break
			}
		}
	}

	return
}

func (h *Header) IndentLevel() int {
	return 0
}

func (h *Header) ChildIndentLevel() int {
	return h.level + 1
}

func (h *Header) Level() int {
	return h.level
}

func (h *Header) SetLevel(level int) {
	h.level = level
}

func (h *Header) AddChild(r Render) error {
	h.children = append(h.children, r)

	return nil
}

func (h *Header) Children() []Render {
	return h.children
}

func (b *Header) ChildrenRec(depth int) (children []Render) {
	if depth == 0 {
		return
	}

	for _, child := range b.Children() {
		children = append(children, child)
		children = append(children, child.ChildrenRec(depth-1)...)
	}

	return
}

func (b *Header) Uid() Uid {
	return NewUid(b.properties.content["ID"].Int().Unwrap())
}

// GetParentUid returns the UID of the parent header, if it exists
func (h *Header) ParentUid() Uid {
	return option.Map(h.Parent, func(r Render) Uid {
		return r.Uid()
	}).UnwrapOr(NewUid(0))
}

// CreateSubheader creates a new header as a child of the current header
// The new header's level will be one more than the parent's level
// A unique UID will be generated and assigned using NewPropertiesWithUID
func (h *Header) CreateSubheader(title string, status HeaderStatus) *Header {
	newHeader := &Header{
		level:    h.Level() + 1,
		status:   status,
		Content:  title,
		Parent:   option.Some(Render(h)),
		children: []Render{},
	}

	newHeader.properties = NewPropertiesWithUID(newHeader)
	h.children = append(h.children, newHeader)

	return newHeader
}

// ToggleCheckboxByIndex finds a bullet at the given index within this header's children
// and toggles its checkbox state. Returns the updated bullet or an error.
func (h *Header) ToggleCheckboxByIndex(index int) (*Bullet, error) {
	if index < 0 || index >= len(h.children) {
		return nil, fmt.Errorf("index %d out of range", index)
	}

	bullet, ok := h.children[index].(*Bullet)
	if !ok {
		return nil, fmt.Errorf("child at index %d is not a bullet", index)
	}

	if !bullet.HasCheckbox() {
		return nil, fmt.Errorf("bullet at index %d does not have a checkbox", index)
	}

	bullet.ToggleCheckbox()
	return bullet, nil
}

// CompleteCheckboxByIndex finds a bullet at the given index within this header's children
// and marks it as completed. Returns the updated bullet or an error.
func (h *Header) CompleteCheckboxByIndex(index int) (*Bullet, error) {
	if index < 0 || index >= len(h.children) {
		return nil, fmt.Errorf("index %d out of range", index)
	}

	bullet, ok := h.children[index].(*Bullet)
	if !ok {
		return nil, fmt.Errorf("child at index %d is not a bullet", index)
	}

	if !bullet.HasCheckbox() {
		return nil, fmt.Errorf("bullet at index %d does not have a checkbox", index)
	}

	bullet.CompleteCheckbox()
	return bullet, nil
}

func (h *Header) Status() RenderStatus {
	return RenderStatus(h.status)
}

func (h *Header) SetStatus(status HeaderStatus) {
	h.status = status
}

func (h *Header) TagList() (list TagList) {
	if parent, ok := h.Parent.Split(); ok {
		list = parent.TagList()
	}

	if tags, ok := h.Tags.Split(); ok {
		list = append(list, tags...)
	}

	return
}

func (h *Header) Preview(length int) string {
	if length < 0 || length >= len(h.Content) {
		return h.Content
	}

	return h.Content[:length]
}

func (h *Header) Path() string {
	if parent, ok := h.Parent.Split(); ok {
		return parent.Path() + "/" + h.Uid().String()
	}

	return h.Uid().String()
}
