package orgmcp

import (
	"bufio"
	"fmt"
	"iter"
	"main/utils/itertools"
	"main/utils/option"
	"main/utils/slice"
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
	case "done":
		return Done
	}

	return None
}

func (s HeaderStatus) toString() string {
	if s == None {
		return ""
	}

	return fmt.Sprintf("%s ", s)
}

const (
	None HeaderStatus = "NONE"
	Todo HeaderStatus = "TODO"
	Next HeaderStatus = "NEXT"
	Prog HeaderStatus = "PROG"
	Done HeaderStatus = "DONE"
)

var SPECIAL_TOKENS = []string{"[", ":"}

// Enforce that Header implements the Render interface at compile time
var _ Render = (*Header)(nil)

type Header struct {
	Status   HeaderStatus
	Level    int
	Progress option.Option[Progress]
	Tags     option.Option[TagList]
	location int

	Parent   option.Option[Render]
	children []Render
	meta     option.Option[Metadata]

	Content string
}

func HeaderFromString(str string, reader *bufio.Reader) option.Option[Header] {
	if !strings.HasPrefix(str, "*") {
		return option.None[Header]()
	}

	// Trim the line to remove any trailing newline characters
	str = strings.TrimSpace(str)

	header := Header{}
	next, stop := iter.Pull(strings.SplitSeq(str, " "))
	defer stop()

	level, _ := next()
	header.Level = strings.Count(level, "*") - 1

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

	indent := header.Level + 2
	meta, err := reader.Peek(indent + 6)

	if err != nil {
		return option.Some(header)
	}

	if slices.ContainsFunc([]MetadataString{CLOSED}, func(m MetadataString) bool {
		return strings.ToUpper(strings.TrimSpace(string(meta))) == string(m)
	}) {
		meta, err = reader.ReadBytes('\n')
		header.meta = option.Some(MetadataFromString(string(meta), &header))
	}

	return option.Some(header)
}

func (h *Header) AddChildren(render ...Render) {
	h.children = append(h.children, render...)
}

func (h *Header) Render(builder *strings.Builder) {
	builder.WriteString(strings.Repeat("*", h.Level+1))
	builder.WriteString(" ")
	builder.WriteString(h.Status.toString())
	builder.WriteString(h.Content)

	h.Progress.Then(func(p Progress) {
		builder.WriteRune(' ')
		p.Render(builder)
	})

	h.Tags.Then(func(tl TagList) {
		builder.WriteRune(' ')
		tl.Render(builder)
	})

	builder.WriteRune('\n')

	h.meta.Then(func(m Metadata) {
		m.Render(builder)
		builder.WriteRune('\n')
	})

	children := itertools.FromSlice(h.children)
	itertools.ForEach(children, func(child Render) {
		child.Render(builder)
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
		}

		return progress
	})
}

func (h *Header) Location() int {
	return h.location
}

func (h *Header) IndentLevel() int {
	return h.Level + 2
}

func (h *Header) AddChild(r Render) error {
	h.children = append(h.children, r)

	return nil
}

func (h *Header) Children() []Render {
	return h.children
}
