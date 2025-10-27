package orgmcp

import (
	"bufio"
	"io"
	"main/utils/itertools"
	"main/utils/logging"
	"main/utils/option"
	"main/utils/result"
	"maps"
	"slices"
	"strings"
)

type OrgFile struct {
	name     string
	children []Render

	items map[int]Render
}

// Enforce that OrgFile implements the Render interface at compile time
var _ Render = (*OrgFile)(nil)

func OrgFileFromReader(reader io.Reader) result.Result[OrgFile] {
	org_file := OrgFile{
		items: map[int]Render{},
	}

	org_file.items[0] = &org_file

	buf_reader := bufio.NewReader(reader)

	var current_parent Render = &org_file
	var current_line = 1

	for val, err := buf_reader.ReadBytes('\n'); true; val, err = buf_reader.ReadBytes('\n') {
		if err == io.EOF {
			break
		}

		if err != nil {
			return result.Err[OrgFile](err)
		}

		if len(val) == 0 {
			continue
		}

		switch val[0] {
		case '*':
			HeaderFromString(string(val), buf_reader).Then(func(h Header) {
				h.Parent = option.Some(current_parent)
				h.location = current_line
				current_parent.AddChild(&h)
				org_file.items[current_line] = &h
				current_line += 1
				current_parent = &h
			})
		case ' ':
			ParseIndentedLine(string(val), current_parent).Then(func(r Render) {
				org_file.items[current_line] = r
				current_line += 1
				current_parent.AddChild(r)
			})
		default:

		}
	}

	return result.Ok(org_file)
}

func ParseIndentedLine(str string, parent Render) option.Option[Render] {
	trimmed := strings.TrimSpace(str)

	switch trimmed[0] {
	case '-':
		fallthrough
	case '*':
		return option.Map(BulletFromString(trimmed, parent), func(b Bullet) Render { return &b })
	}

	return logging.TODO[option.Option[Render]]()
}

func (of *OrgFile) Render(builder *strings.Builder) {
	for _, child := range of.children {
		child.Render(builder)
	}
}

func (of *OrgFile) Location() int {
	return 0
}

func (of *OrgFile) CheckProgress() option.Option[Progress] {
	return option.None[Progress]()
}

func (of *OrgFile) IndentLevel() int {
	return 0
}

func (of *OrgFile) Children() []Render {
	return of.children
}

func (h *OrgFile) AddChild(r Render) error {
	h.children = append(h.children, r)

	return nil
}

func (h *OrgFile) GetLine(line int) option.Option[Render] {
	if item := h.items[line]; item != nil {
		return option.Some(item)
	}

	return option.None[Render]()
}

func (h *OrgFile) GetTag(tag string) option.Option[*Header] {
	return option.Cast[Render, *Header](
		itertools.Find(
			maps.Values(h.items),
			func(r Render) bool {
				header, ok := r.(*Header)
				if !ok {
					return false
				}

				return option.Map(header.Tags, func(t TagList) bool {
					return slices.Contains(t, tag)
				}).UnwrapOr(false)
			},
		),
	)
}

func (h *OrgFile) GetHeaderByStatus(status HeaderStatus) []*Header {
	headers := []*Header{}

	for _, child := range h.children {
		if header, ok := child.(*Header); ok {
			headers = GetHeaderRec(header, func(h *Header) bool {
				return h.Status == status
			}, headers)
		}
	}

	return headers
}

func GetHeaderRec(header *Header, predicate func(*Header) bool, headers []*Header) []*Header {
	if predicate(header) {
		headers = append(headers, header)
		return headers
	}

	for _, child := range header.children {
		if header, ok := child.(*Header); ok {
			headers = GetHeaderRec(header, predicate, headers)
		}
	}

	return headers
}
