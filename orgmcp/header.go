package orgmcp

import (
	"fmt"
	"github.com/p3rtang/org-mcp/embeddings"
	"github.com/p3rtang/org-mcp/utils/itertools"
	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/reader"
	"github.com/p3rtang/org-mcp/utils/slice"
	"iter"
	"slices"
	"strings"
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
	Status   HeaderStatus
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
	header.level = strings.Count(level, "*") - 1

	part, end := next()

	header.Status = StatusFromString(part)

	if header.Status != None {
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

	return nil
}

func (b *Header) RemoveChildren(uids ...Uid) error {
	b.children = slice.Filter(b.children, func(r Render) bool {
		return !slices.Contains(uids, r.Uid())
	})

	return nil
}

func (h *Header) Render(builder *strings.Builder, depth int) {
	builder.WriteString(strings.Repeat("*", h.Level()+1))
	builder.WriteString(" ")
	if h.Status != None {
		builder.WriteString(h.Status.String())
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
	if h.Progress.IsNone() && h.Status != None {
		return option.Some(Progress{done: option.Some(h.Status == Done)})
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

		if h.Progress.AndThen(func(p Progress) bool { return p.Done() }) && h.Status != None {
			h.Status = Done
		} else if h.Progress.AndThen(func(p Progress) bool { return p.Prog() }) && h.Status != None && h.Status != Done {
			h.Status = Prog
		}

		return progress
	})
}

func (h *Header) Location() int {
	return h.location
}

func (h *Header) IndentLevel() int {
	return h.level + 2
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

func (b *Header) ChildrenRec() []Render {
	children := []Render{}

	for _, child := range b.Children() {
		children = append(children, child.ChildrenRec()...)
	}

	return children
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
		level:      h.Level() + 1,
		Status:     status,
		Content:    title,
		Parent:     option.Some(Render(h)),
		children:   []Render{},
		properties: NewPropertiesWithUID(),
	}

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
